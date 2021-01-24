package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-various/ginplus"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport/logical"
	"github.com/go-various/pool"
	"github.com/hashicorp/go-hclog"
	"strings"
	"time"
)

var _ logical.Transport = (*Transport)(nil)

type Transport struct {
	pm          *pluginregister.PluginManager
	logger      hclog.Logger
	workerPool  *pool.WorkerPool
	ctx         context.Context
	cancel      context.CancelFunc
	ginServer   *ginplus.Server
	running     chan bool
	workerSize  int
	authMethod  *logical.Method
	authEnabled bool
	security    logical.Security
}

func NewHTTPTransport(m *pluginregister.PluginManager, workerSize int, logger hclog.Logger) *Transport {
	ctx, cancel := context.WithCancel(context.Background())
	gw := &Transport{
		pm:          m,
		logger:      logger.Named("http-transport"),
		ctx:         ctx,
		cancel:      cancel,
		running:     make(chan bool, 1),
		workerSize:  workerSize,
		ginServer:   ginplus.NewServer(),
		authEnabled: true,
	}
	return gw
}

func (m *Transport) SetAuthEnabled() {
	m.authEnabled = true
}

func (m *Transport) SetAuthDisabled() {
	m.authEnabled = false
}

func (m *Transport) SetSecurity(security logical.Security) {
	m.security = security
}

func (m *Transport) SetAuthMethod(method string) error {
	if method == "" {
		return nil
	}
	methods := strings.Split(method, ".")[:]
	if len(methods) != 3 {
		return errors.New("auth method error")
	}
	m.authMethod = &logical.Method{
		Backend:   methods[0],
		Namespace: methods[1],
		Operation: methods[2],
	}
	return nil
}

//关闭网关
func (m *Transport) Shutdown() {
	defer func() {
		if m.logger.IsTrace() {
			m.logger.Trace("exited")
		}
	}()

	m.ginServer.Shutdown(context.Background())

	m.workerPool.Shutdown()

	select {
	case <-m.workerPool.Running():
	case <-time.After(time.Second * 1):
	}
	close(m.running)
}

//网关是否在运行(阻塞等待)
func (m *Transport) Running() <-chan bool {
	return m.running
}

func (m *Transport) Router() *ginplus.Server {
	return m.ginServer
}

func (m *Transport) AddRouter(method, router string, handleFunc func(*ginplus.Context) error) {
	m.ginServer.Router.Handle(method, router, func(c *gin.Context) {
		ctx := new(ginplus.Context)
		ctx.Context = c
		err := handleFunc(ctx)
		if nil != err {
			m.logger.Error("handle", "method", method, "router", router, err)
			c.AbortWithError(200, err)
		}
		c.Next()
	})
}

func (m *Transport) Listen(addr string, port uint) error {
	if err := m.ginServer.Listen(addr, port); err != nil {
		return err
	}
	return nil
}

func (m *Transport) Serve(basePath string) error {
	m.startWorkerPool(m.workerSize)
	m.api(basePath)
	m.open(basePath)
	return m.ginServer.Serve()
}
