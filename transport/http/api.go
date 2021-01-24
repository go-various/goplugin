package http

import (
	"github.com/gin-gonic/gin"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/goplugin/transport/logical"
	"github.com/go-various/helper/jsonutil"
	"github.com/go-various/pool"
	"github.com/google/uuid"
	"strings"
)

func (m *Transport) open(basePath string) {

	m.engine.POST(basePath+"/open", func(c *gin.Context) {
		request := new(transport.Request)
		if err := c.ShouldBindJSON(request); err != nil {
			c.SecureJSON(200, transport.Error(transport.ReplyCodeFailure, err.Error()))
			return
		}
		methods := strings.Split(request.Method, ".")[:]
		if len(methods) != 3 {
			c.SecureJSON(200,
				transport.Error(transport.ReplyCodeMethodInvalid, "method error"))
			return
		}
		method := transport.Method{
			Backend:   methods[0],
			Namespace: methods[1],
			Operation: methods[2],
		}

		if m.Security != nil {
			client := &transport.Client{
				RemoteAddr: GetRemoteAddr(c),
				Referer:    c.Request.Referer(),
				UserAgent:  c.Request.UserAgent(),
			}
			if err := m.Security.Blocker(client); err != nil {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeReqBlocked, err.Error()))
				return
			}

			if err := m.Security.RateLimiter(client); err != nil {
				c.SecureJSON(200, transport.Error(transport.ReplyCodeRateLimited, err.Error()))
				return
			}
		}
		bs, _ := jsonutil.EncodeJSON(request)
		m.invokeRequest(c, method, string(bs))
	})
}
func (m *Transport) api(basePath string) {
	m.engine.POST(basePath+"/api", func(c *gin.Context) {
		request := new(transport.Request)
		if err := c.ShouldBindJSON(request); err != nil {
			c.SecureJSON(200, transport.Error(transport.ReplyCodeFailure, err.Error()))
			return
		}
		methods := strings.Split(request.Method, ".")[:]
		if len(methods) != 3 {
			c.SecureJSON(200,
				transport.Error(transport.ReplyCodeMethodInvalid, "method error"))
			return
		}
		method := transport.Method{
			Backend:   methods[0],
			Namespace: methods[1],
			Operation: methods[2],
		}

		if m.Security != nil {
			if !m.Security.SignVerify(request) {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeSignInvalid, "invalid sign"))
				return
			}
			client := &transport.Client{
				RemoteAddr: GetRemoteAddr(c),
				Referer:    c.Request.Referer(),
				UserAgent:  c.Request.UserAgent(),
			}
			if err := m.Security.Blocker(client); err != nil {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeReqBlocked, err.Error()))
				return
			}

			if err := m.Security.RateLimiter(client); err != nil {
				c.SecureJSON(200, transport.Error(transport.ReplyCodeRateLimited, err.Error()))
				return
			}
		}
		m.invokeRequest(c, method, request.Data)
	})
}

func (m *Transport) invokeRequest(c *gin.Context, method transport.Method, data string) {
	request := &gological.Request{
		ID:        uuid.New().String(),
		Namespace: method.Namespace,
		Operation: method.Operation,
		Data:      []byte(data),
		Headers:   c.Request.Header,
		Token:     c.Request.Header.Get(gological.AuthTokenName),
		Connection: &gological.Connection{
			RemoteAddr: GetRemoteAddr(c),
			ConnState:  c.Request.TLS,
		},
	}
	workerData := &logical.WorkerData{
		Backend: method.Backend,
		Request: request,
	}
	output := make(chan *logical.WorkerReply, 1)
	subject := pool.NewSubject(workerData)

	subject.Observer(m.NewObserver(output))

	m.WorkerPool.Input(subject)

	select {
	case d := <-output:
		m.writerReply(c, d)
	}
	close(output)
}

func GetRemoteAddr(c *gin.Context) string {
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr == "127.0.0.1" {
		remoteAddr = c.GetHeader("X-Forwarded-For")
	}
	return remoteAddr
}
