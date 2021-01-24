package rpc

import (
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/goplugin/transport/logical"
	"github.com/go-various/pool"
	"github.com/google/uuid"
	"net/rpc"
	"strings"
)


func (t *Transport) invokeRequest(args *transport.Request, reply *transport.Response) {
	methods := strings.Split(args.Method, ".")[:]
	if len(methods) != 3 {
		reply.Code = transport.ReplyCodeMethodInvalid
		reply.Message = "method error"
		return
	}
	method := transport.Method{
		Backend:   methods[0],
		Namespace: methods[1],
		Operation: methods[2],
	}

	request := &gological.Request{
		ID:        uuid.New().String(),
		Namespace: method.Namespace,
		Operation: method.Operation,
		Data:      []byte(args.Data),
		Connection: &gological.Connection{
		},
	}
	workerData := &logical.WorkerData{
		Backend: method.Backend,
		Request: request,
	}
	output := make(chan *logical.WorkerReply, 1)
	subject := pool.NewSubject(workerData)

	subject.Observer(t.NewObserver(output))

	t.WorkerPool.Input(subject)

	select {
	case d := <-output:
		if d.Err != nil{
			reply.Code = transport.ReplyCodeFailure
			reply.Message = d.Err.Error()
		}else {
			reply.Result = d.Result
		}
	}
	close(output)
}

type Service struct {
	trans *Transport
	rpc *rpc.Server
}

func (t *Service) Invoke(args *transport.Request, reply *transport.Response)error  {
	 t.trans.invokeRequest(args, reply)
	 return nil
}