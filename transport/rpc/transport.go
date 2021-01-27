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

var handle = &codec.MsgpackHandle{}

func (m *Transport) handleConn(conn net.Conn) {
	m.Logger.Trace("connected", "remote", conn.RemoteAddr())
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
	defer func() {
		metrics.IncrCounter([]string{"transport", "rpc", "success"}, 1)
		m.Logger.Trace("closed", "remote", conn.RemoteAddr())
		conn.Close()
	}()
	rpcCodec := msgpackrpc.NewCodecFromHandle(true, true, conn, handle)
	for {
		select {
		case <-m.Ctx.Done():
			return
		default:
		}
		if err := m.rpcServer.ServeRequest(rpcCodec); err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "closed") {
				m.Logger.Error("RPC error", "conn", conn.RemoteAddr(), "error", err)
				metrics.IncrCounter([]string{"transport", "rpc", "error"}, 1)
			}
			return
		}
	}

}
