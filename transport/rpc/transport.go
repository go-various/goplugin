package rpc

import (
	"fmt"
	"github.com/armon/go-metrics"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/goplugin/transport/logical"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-msgpack/codec"
	msgpackrpc "github.com/hashicorp/net-rpc-msgpackrpc"
	"io"
	"net"
	"net/rpc"
	"strings"
)

var _ transport.Transport = (*Transport)(nil)

type Transport struct {
	*logical.Transport
	rpcServer *rpc.Server
	listener  net.Listener
}

func (t *Transport) AddHandle(handle interface{}, args... string)error {
	if len(args) == 1{
		return t.rpcServer.RegisterName(args[0], handle)
	}else {
		return t.rpcServer.Register(handle)
	}
}

func NewTransport(m *pluginregister.PluginManager,
	workerSize int, logger hclog.Logger) *Transport {
	trans := logical.New(logger.Named("rpc"), workerSize)
	trans.PluginManager = m
	return &Transport{
		Transport: trans,
		rpcServer: rpc.NewServer(),
	}
}

func (t *Transport) Listen(addr string, port uint) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	t.listener = ln
	return err
}

func (t *Transport) Start() error {
	t.StartWorkerPool()

	if err := t.AddHandle( &Service{trans: t, rpc: t.rpcServer}, "Transport"); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-t.Ctx.Done():
				return
			default:
			}
			conn, err := t.listener.Accept()
			if err != nil {
				t.Logger.Error("rpc.Start: accept:", "err", err.Error())
				return
			}

			t.handleConn(conn)
		}
	}()
	return nil
}

func (t *Transport) handleConn(conn net.Conn) {
	if t.Security != nil {
		client := &transport.Client{
			RemoteAddr: conn.RemoteAddr().String(),
		}
		if err := t.Security.Blocker(client); err != nil {
			conn.Close()
			return
		}
		if err := t.Security.RateLimiter(client); err != nil {
			conn.Close()
			return
		}
	}

	rpcCodec := msgpackrpc.NewCodecFromHandle(true, true, conn, &codec.MsgpackHandle{})
	if err := t.rpcServer.ServeRequest(rpcCodec); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "closed") {
			t.Logger.Error("RPC error",
				"conn", conn.RemoteAddr(),
				"error", err,
			)
			metrics.IncrCounter([]string{"rpc", "request_error"}, 1)
		}
		return
	}
	metrics.IncrCounter([]string{"rpc", "request"}, 1)
}
