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

func (m *Transport) AddHandle(handle interface{}, args ...string) error {
	if len(args) == 1 {
		return m.rpcServer.RegisterName(args[0], handle)
	} else {
		return m.rpcServer.Register(handle)
	}
}

func NewTransport(m *pluginregister.PluginManager,
	workerSize int, logger hclog.Logger) *Transport {
	trans := logical.New(logger.Named("rpc-transport"), workerSize)
	trans.PluginManager = m
	return &Transport{
		Transport: trans,
		rpcServer: rpc.NewServer(),
	}
}

func (m *Transport) Listen(addr string, port uint) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	m.listener = ln
	return err
}

func (m *Transport) Start() error {
	m.StartWorkerPool()

	if err := m.AddHandle(&Service{trans: m, rpc: m.rpcServer}, "Transport"); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-m.Ctx.Done():
				return
			default:
			}
			conn, err := m.listener.Accept()
			if err != nil {
				m.Logger.Error("rpc.Start: accept:", "err", err.Error())
				return
			}

			m.handleConn(conn)
		}
	}()
	return nil
}

func (m *Transport) handleConn(conn net.Conn) {
	if m.Security != nil {
		client := &transport.Client{
			RemoteAddr: conn.RemoteAddr().String(),
		}
		if err := m.Security.Blocker(client); err != nil {
			conn.Close()
			return
		}
		if err := m.Security.RateLimiter(client); err != nil {
			conn.Close()
			return
		}
	}
	rpcCodec := msgpackrpc.NewCodecFromHandle(true, true, conn, &codec.MsgpackHandle{})
	if err := m.rpcServer.ServeRequest(rpcCodec); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "closed") {
			m.Logger.Error("RPC error", "conn", conn.RemoteAddr(), "error", err)
			metrics.IncrCounter([]string{"rpc", "request_error"}, 1)
		}
		return
	}

	metrics.IncrCounter([]string{"rpc", "request"}, 1)

}
