/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package subscriber

import (
	"fmt"
	"sync"
	"testing"
)

const LEN = 3

type eventSubTest struct {
	wg              sync.WaitGroup
	bytesCh         chan []byte
	bytesSub        Subscription
	bytesSubscribed chan bool
	bytesFeed       Feed
	stringCh        []chan string
	stringSub       []Subscription
	strSubscribed   []chan bool
	stringFeed      Feed
}

func (s *eventSubTest) SendStrMsg(msg string) {
	s.stringFeed.Send(msg)
	fmt.Printf("send string msg: %s\n", msg)
}

func (s *eventSubTest) SendBytesMsg(msg []byte) {
	s.bytesFeed.Send(msg)
	fmt.Printf("send bytes msg: %s\n", string(msg))
}

func (s *eventSubTest) SubscribeStringEvent(ch chan<- string, index int) {
	s.stringSub[index] = s.stringFeed.Subscribe(ch)
	s.strSubscribed[index] <- true
	fmt.Printf("subscribe string[%v] event: %v\n", index, s.stringSub)
}

func (s *eventSubTest) SubscribeBytesEvent(ch chan<- []byte) {
	s.bytesSub = s.bytesFeed.Subscribe(ch)
	s.bytesSubscribed <- true
	fmt.Printf("subscribe bytes event: %v\n", s.bytesSub)
}

func TestFeed(t *testing.T) {
	et := &eventSubTest{
		stringFeed:      Feed{},
		bytesFeed:       Feed{},
		stringCh:        make([]chan string, LEN),
		bytesCh:         make(chan []byte),
		strSubscribed:   make([]chan bool, LEN),
		bytesSubscribed: make(chan bool),
		stringSub:       make([]Subscription, LEN),
	}

	for i := 0; i < LEN; i++ {
		et.stringCh[i] = make(chan string)
		et.strSubscribed[i] = make(chan bool)
	}

	et.wg.Add(1)
	go func() {
		cnt := 0
		for {
			select {
			case x0 := <-et.stringCh[0]:
				fmt.Printf("string chan0 read data[%s]\n", x0)
				cnt++
			case x1 := <-et.stringCh[1]:
				fmt.Printf("string chan1 read data[%s]\n", x1)
				cnt++
			case x2 := <-et.stringCh[2]:
				fmt.Printf("string chan2 read data[%s]\n", x2)
				cnt++
			case y0 := <-et.bytesCh:
				fmt.Printf("bytes chan read data[%s]\n", string(y0))
				et.bytesSub.Unsubscribe()
				cnt++
			default:

			}

			if cnt == LEN+1 {
				for i := 0; i < LEN; i++ {
					et.stringSub[i].Unsubscribe()
				}
				fmt.Printf("unsubscribe select, cnt:%v\n", cnt)
				break
			}
		}
		et.wg.Done()
	}()

	et.wg.Add(1)
	go func() {
		cnt := 0
		subedCnt := 0
		for {
			select {
			case <-et.strSubscribed[0]:
				fmt.Println("receive subscribed string chan0")
				subedCnt++
			case <-et.strSubscribed[1]:
				fmt.Println("receive subscribed string chan1")
				subedCnt++
			case <-et.strSubscribed[2]:
				fmt.Println("receive subscribed string chan2")
				subedCnt++
			case <-et.bytesSubscribed:
				et.SendBytesMsg([]byte("test bytes feed"))
				cnt++
			default:

			}

			if subedCnt == LEN {
				et.SendStrMsg("test string feed")
				subedCnt = 0
				cnt++
			}

			if cnt == 2 {
				break
			}
		}
		et.wg.Done()
	}()

	et.wg.Add(1)
	go func() {
		et.SubscribeStringEvent(et.stringCh[0], 0)
		et.SubscribeStringEvent(et.stringCh[1], 1)
		et.SubscribeStringEvent(et.stringCh[2], 2)
		et.wg.Done()
	}()

	et.wg.Add(1)
	go func() {
		et.SubscribeBytesEvent(et.bytesCh)
		et.wg.Done()
	}()

	et.wg.Wait()
	fmt.Printf("chan cnt of string: %v, bytes: %v\n", len(et.stringCh), len(et.bytesCh))
}
