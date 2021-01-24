package rpc

import (
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-msgpack/codec"
	msgpackrpc "github.com/hashicorp/net-rpc-msgpackrpc"
	"net"
	"testing"
)

func TestNewTransport(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:              "rpc",
		Level:             hclog.Trace,
		IncludeLocation:   true,
	})
	pm := pluginregister.NewPluginManager("plugin",nil, nil, logger)
	trans := NewTransport(pm, 8, logger)
	if err := trans.Listen("127.0.0.1", 6000); err != nil {
		t.Fatal(err)
		return
	}
	if err := trans.Start(); err != nil {
		t.Fatal(err)
		return
	}
	done := make(chan byte)
	<-done
}

func TestService_Invoke(t *testing.T) {
	//RPC Communication (client side)
	conn, err := net.Dial("tcp", "127.0.0.1:6000")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	//t.Log("connected", conn)

	//t.Log("write", conn)
	//conn.SetDeadline(time.Now().Add(time.Second * 5))
	//rpcCodec := msgpackrpc.NewCodecFromHandle(true, true, conn, pool.MsgpackHandle)
	//rpcCodec := codec.MsgpackSpecRpc.ClientCodec(conn, pool.MsgpackHandle)
	args := transport.Request{
		Method:    "account.user.login",
		Version:   "",
		Timestamp: "",
		SignType:  "",
		Sign:      "",
		Data:      "",
	}
	var reply transport.Response
	rpcCodec := msgpackrpc.NewCodecFromHandle(true, true, conn, &codec.MsgpackHandle{})
	if err := msgpackrpc.CallWithCodec(rpcCodec,"Transport.Invoke", args, &reply); err != nil {
		t.Error(err)
		return
	}
	t.Log(reply)
}
