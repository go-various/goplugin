package http

import (
	"github.com/go-various/goplugin/pluginregister"
	"github.com/hashicorp/go-hclog"
	"testing"
)

func TestNewTransport(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:              "rpc",
		Level:             hclog.Trace,
		IncludeLocation:   true,
	})
	pm := pluginregister.NewPluginManager("plugin",nil, nil, logger)
	trans := NewTransport(pm, "",8, logger)
	if err := trans.Listen("127.0.0.1", 7000); err != nil {
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