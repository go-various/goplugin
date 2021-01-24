package http

import (
	"github.com/gin-gonic/gin"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/helper/jsonutil"
	"github.com/go-various/pool"
	"github.com/google/uuid"
	"strings"
)

func (m *Transport) endpoints(basePath string) {
}
func (m *Transport) open(basePath string) {
	m.ginServer.Router.POST(basePath+"/open", func(c *gin.Context) {
		request := new(transport.RequestArgs)
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

		if m.security != nil {
			client := &transport.Client{
				RemoteAddr: GetRemoteAddr(c),
				Referer:    c.Request.Referer(),
				UserAgent:  c.Request.UserAgent(),
			}
			if err := m.security.Blocker(&method, client); err != nil {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeReqBlocked, err.Error()))
				return
			}

			if err := m.security.RateLimiter(&method, client); err != nil {
				c.SecureJSON(200, transport.Error(transport.ReplyCodeRateLimited, err.Error()))
				return
			}
		}
		bs, _ := jsonutil.EncodeJSON(request)
		m.invokeRequest(c, method, string(bs))
	})
}
func (m *Transport) api(basePath string) {
	m.ginServer.Router.POST(basePath+"/api", func(c *gin.Context) {
		request := new(transport.RequestArgs)
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

		if m.security != nil {
			if !m.security.SignVerify(request) {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeSignInvalid, "invalid sign"))
				return
			}
			client := &transport.Client{
				RemoteAddr: GetRemoteAddr(c),
				Referer:    c.Request.Referer(),
				UserAgent:  c.Request.UserAgent(),
			}
			if err := m.security.Blocker(&method, client); err != nil {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeReqBlocked, err.Error()))
				return
			}

			if err := m.security.RateLimiter(&method, client); err != nil {
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
	workerData := &workData{
		backend: method.Backend,
		request: request,
	}
	output := make(chan *workerReply, 1)
	subject := pool.NewSubject(workerData)

	subject.Observer(m.NewObserver(output))

	m.workerPool.Input(subject)

	select {
	case d := <-output:
		m.writerReply(c, d)
	}
}

func GetRemoteAddr(c *gin.Context) string {
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr == "127.0.0.1" {
		remoteAddr = c.GetHeader("X-Forwarded-For")
	}
	return remoteAddr
}
