package xvm

import (
	"fmt"

	"chainmaker.org/chainmaker-go/common/serialize"
	"chainmaker.org/chainmaker-go/wxvm/xvm/exec"
)

const (
	contextIDKey = "ctxid"
	responseKey  = "callResponse"
)

type responseDesc struct {
	Body  []byte
	Error bool
}

type contextServiceResolver struct {
	contextService *ContextService
}

func NewContextServiceResolver(service *ContextService) exec.Resolver {
	return &contextServiceResolver{
		contextService: service,
	}
}

func (s *contextServiceResolver) ResolveGlobal(module, name string) (int64, bool) {
	return 0, false
}

func (s *contextServiceResolver) ResolveFunc(module, name string) (interface{}, bool) {
	fullname := module + "." + name
	switch fullname {
	case "env._call_method":
		return s.cCallMethod, true
	case "env._fetch_response":
		return s.cFetchResponse, true
	default:
		return nil, false
	}
}
func (s *contextServiceResolver) cFetchResponse(ctx exec.Context, userBuf, userLen uint32) uint32 {
	codec := exec.NewCodec(ctx)
	iresponse := ctx.GetUserData(responseKey)
	if iresponse == nil {
		exec.Throw(exec.NewTrap("call fetchResponse on nil value"))
	}
	response := iresponse.(responseDesc)
	userbuf := codec.Bytes(userBuf, userLen)
	if len(response.Body) != len(userbuf) {
		exec.Throw(exec.NewTrap(fmt.Sprintf("call fetchResponse with bad length, got %d, expect %d", len(userbuf), len(response.Body))))
	}
	copy(userbuf, response.Body)
	success := uint32(1)
	if response.Error {
		success = 0
	}
	ctx.SetUserData(responseKey, nil)
	return success
}

func (s *contextServiceResolver) cCallMethod(
	ctx exec.Context,
	methodAddr, methodLen uint32,
	requestAddr, requestLen uint32,
	responseAddr, responseLen uint32,
	successAddr uint32) uint32 {

	codec := exec.NewCodec(ctx)
	ctxId := ctx.GetUserData(contextIDKey).(int64)
	method := codec.String(methodAddr, methodLen)
	requestBuf := codec.Bytes(requestAddr, requestLen)
	responseBuf := codec.Bytes(responseAddr, responseLen)

	var respMessage []*serialize.EasyCodecItem
	var err error

	switch method {
	case "GetObject":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.GetObject(ctxId, reqItems)
	case "PutObject":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.PutObject(ctxId, reqItems)
	case "DeleteObject":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.DeleteObject(ctxId, reqItems)
	case "NewIterator":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.NewIterator(ctxId, reqItems)
	case "GetCallArgs":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.GetCallArgs(ctxId, reqItems)
	case "SetOutput":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.SetOutput(ctxId, reqItems)
	case "ContractCall":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.ContractCall(ctxId, reqItems, ctx.GasUsed())
	case "LogMsg":
		reqItems := serialize.EasyUnmarshal(requestBuf)
		respMessage, err = s.contextService.LogMsg(ctxId, reqItems)
	default:
		s.contextService.logger.Errorw("no such method ", method)
	}
	if err != nil {
		s.contextService.logger.Errorw("failed to call method:", method, err)
		codec.SetUint32(successAddr, 1)
		return uint32(0)
	}

	possibleResponseBuf := serialize.EasyMarshal(respMessage)

	// fast path
	if err != nil {
		s.contextService.logger.Errorw("contract syscall error", "ctxid", ctxId, "method", method, "error", err)
		msg := err.Error()
		if len(msg) <= len(responseBuf) {
			copy(responseBuf, msg)
			codec.SetUint32(successAddr, 0)
			return uint32(len(msg))
		}
	} else {
		if len(possibleResponseBuf) <= len(responseBuf) {
			copy(responseBuf, possibleResponseBuf)
			codec.SetUint32(successAddr, 1)
			return uint32(len(possibleResponseBuf))
		}
	}

	// slow path
	var responseDesc responseDesc
	if err != nil {
		s.contextService.logger.Errorw("contract service call error", "ctxid", ctxId, "method", method, "error", err)
		responseDesc.Error = true
		responseDesc.Body = []byte(err.Error())
	} else {
		responseDesc.Body = possibleResponseBuf
	}
	ctx.SetUserData(responseKey, responseDesc)
	return uint32(len(responseDesc.Body))
}