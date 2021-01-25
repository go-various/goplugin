package rpc

import (
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/transport"
	"github.com/google/uuid"
	"net/rpc"
	"strings"
)

func (m *Transport) invokeRequest(args *transport.Request, reply *transport.Response) {
	methods := strings.Split(args.Method, ".")[:]
	if len(methods) != 3 {
		reply.Code = transport.ReplyCodeMethodInvalid
		reply.Message = "method error"
		return
	}
	if m.Security != nil {
		if !m.Security.SignVerify(args) {
			reply.Code = transport.ReplyCodeSignInvalid
			reply.Message = "invalid sign"
			return
		}
	}
	method := transport.Method{
		Backend:   methods[0],
		Namespace: methods[1],
		Operation: methods[2],
	}

	request := &gological.Request{
		ID:         uuid.New().String(),
		Namespace:  method.Namespace,
		Operation:  method.Operation,
		Data:       []byte(args.Data),
		Connection: &gological.Connection{},
	}
	resp := m.Transport.Invoke(method.Backend, request)
	if resp.Err != nil {
		reply.Code = transport.ReplyCodeFailure
		reply.Message = resp.Err.Error()
	} else {
		reply.Result = resp.Result
	}
}

type Service struct {
	trans *Transport
	rpc   *rpc.Server
}

func (t *Service) Invoke(args *transport.Request, reply *transport.Response) error {
	t.trans.invokeRequest(args, reply)
	return nil
}
