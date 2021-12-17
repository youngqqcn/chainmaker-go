/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package subscriber

import (
	"errors"
	"reflect"
	"sync"
)

type feedTypeErr struct {
	got, want reflect.Type
	op        string
}

type Subscription interface {
	Err() <-chan error // returns the error channel
	Unsubscribe()
}

type subImpl struct {
	feed    *Feed
	channel reflect.Value //chan<-
	once    sync.Once
	err     chan error
}

func (s *subImpl) Unsubscribe() {
	s.once.Do(func() {
		s.feed.removeSub(s)
		close(s.err)
	})
}

func (s *subImpl) Err() <-chan error {
	return s.err
}

type Feed struct {
	lock    sync.Mutex
	subList []reflect.SelectCase
	subType reflect.Type
}

func (f *Feed) findSub(data interface{}) int {
	for i, c := range f.subList {
		if c.Chan.Interface() == data {
			return i
		}
	}

	return -1

}

func (f *Feed) deactivateSub(list []reflect.SelectCase, index int) []reflect.SelectCase {
	last := len(list) - 1
	list[index], list[last] = list[last], list[index]
	return list[:last]
}

func (f *Feed) typeCheck(t reflect.Type) bool {
	if f.subType == nil {
		f.subType = t
		return true
	}

	return f.subType == t
}

func (f *Feed) Subscribe(channel interface{}) Subscription {
	val := reflect.ValueOf(channel)
	typ := val.Type()
	if typ.Kind() != reflect.Chan || (typ.ChanDir()&reflect.SendDir == 0) {
		panic(errors.New("parameter is not send channel type"))

	}

	sub := &subImpl{feed: f, channel: val, err: make(chan error, 1)}

	f.lock.Lock()
	defer f.lock.Unlock()

	if !f.typeCheck(typ.Elem()) {
		panic(feedTypeErr{op: "Subscribe", got: typ, want: reflect.ChanOf(reflect.SendDir, f.subType)})
	}

	c := reflect.SelectCase{Dir: reflect.SelectSend, Chan: val}
	f.subList = append(f.subList, c)

	return sub
}

func (f *Feed) removeSub(sub *subImpl) {
	ch := sub.channel.Interface()

	f.lock.Lock()
	defer f.lock.Unlock()

	index := f.findSub(ch)
	f.subList = append(f.subList[:index], f.subList[index+1:]...)
}

func (f *Feed) Send(data interface{}) (cnt int) {
	val := reflect.ValueOf(data)
	if !f.typeCheck(val.Type()) {
		panic(feedTypeErr{op: "Send", got: val.Type(), want: f.subType})
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	subs := f.subList
	for i := 0; i < len(subs); {
		if subs[i].Chan.TrySend(val) {
			subs = f.deactivateSub(subs, i)
			cnt++
		} else {
			i++
		}
	}

	if len(subs) == 0 {
		return cnt
	}

	for i := 0; i < len(subs); i++ {
		subs[i].Send = val
	}

	for {
		index, _, _ := reflect.Select(subs)
		subs = f.deactivateSub(subs, index)
		cnt++

		if len(subs) == 0 {
			return cnt
		}
	}
}
