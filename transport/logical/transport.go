package logical

import (
	"context"
	"errors"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/pool"
	"github.com/hashicorp/go-hclog"
	"strings"
	"time"
)

type Transport struct {
	PluginManager *pluginregister.PluginManager
	Logger        hclog.Logger
	WorkerPool    *pool.WorkerPool
	Ctx           context.Context
	Cancel        context.CancelFunc
	running       chan bool
	workerSize    int
	authMethod    *transport.Method
	authEnabled   bool
	Security      transport.Security
}

func New(logger hclog.Logger, workerSize int) *Transport {
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.Named("transport")
	return &Transport{
		WorkerPool:  pool.NewWorkerPool("transport", context.Background(), log),
		Ctx:         ctx,
		Cancel:      cancel,
		Logger: logger,
		running:     make(chan bool),
		workerSize:  workerSize,
		authEnabled: true,
	}
}

//关闭网关
func (m *Transport) Shutdown() {

	defer func() {
		if m.Logger.IsTrace() {
			m.Logger.Trace("exited")
		}
	}()
	if m.Cancel != nil{
		m.Cancel()
	}
	m.WorkerPool.Shutdown()

	select {
	case <-m.WorkerPool.Running():
	case <-time.After(time.Second * 1):
	}
	close(m.running)
}
func (m *Transport) Running() <-chan bool {
	return m.running
}

func (m *Transport) SetAuthEnabled() {
	m.authEnabled = true
}

func (m *Transport) SetAuthDisabled() {
	m.authEnabled = false
}

func (m *Transport) SetSecurity(security transport.Security) {
	m.Security = security
}

func (m *Transport) SetAuthMethod(method string) error {
	if method == "" {
		return errors.New("auth method can not be empty")
	}
	methods := strings.Split(method, ".")[:]
	if len(methods) != 3 {
		return errors.New("auth method error")
	}
	m.authMethod = &transport.Method{
		Backend:   methods[0],
		Namespace: methods[1],
		Operation: methods[2],
	}
	return nil
}
