package http

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/goplugin/transport/logical"
	"github.com/hashicorp/go-hclog"
	"net"
	"net/http"
)

var _ transport.Transport = (*Transport)(nil)

type Handle func(c *gin.Context)
type Transport struct {
	*logical.Transport
	basePath    string
	listener  net.Listener
	engine  *gin.Engine
	srv *http.Server
}

// AddHandle
// gin.HandlerFunc
// gin.HandlerFunc, httpMethod, relativePath
func (m *Transport) AddHandle(handle interface{}, args... string) error {
	h := handle.(gin.HandlerFunc)
	if len(args) == 0{
		m.engine.Use(h)
	}else if len(args)== 3{
		m.engine.Handle(args[0],args[1], h)
	}else {
		return errors.New("invalid args")
	}
	return nil
}

func NewTransport(m *pluginregister.PluginManager,
	basePath    string,
	workerSize int, logger hclog.Logger) *Transport {
	trans := logical.New(logger.Named("http"), workerSize)
	trans.PluginManager = m
	engine := gin.New()
	engine.Use(gin.Recovery())
	return &Transport{
		Transport: trans,
		basePath: basePath,
		engine: engine,
	}
}

//关闭网关
func (m *Transport) Shutdown() {
	m.Transport.Shutdown()
	m.srv.Shutdown(m.Transport.Ctx)
}

func (m *Transport) Listen(addr string, port uint) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	m.listener = ln
	return err
}

func (m *Transport) Start() error {
	m.StartWorkerPool()
	m.api(m.basePath)
	m.open(m.basePath)
	return m.Serve()
}

func (m *Transport) Serve() (err error) {
	srv := &http.Server{
		Addr:    m.listener.Addr().String(),
		Handler: m.engine,
	}
	m.srv = srv
	go func() {
		err = srv.Serve(m.listener)
	}()
	return err
}