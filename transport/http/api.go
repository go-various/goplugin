package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/helper/jsonutil"
	"github.com/google/uuid"
	"strings"
)

func (m *Transport) getMethod(metStr string) (*transport.Method, error) {
	methods := strings.Split(metStr, ".")[:]
	if len(methods) != 3 {
		return nil, errors.New("method error")
	}
	method := transport.Method{
		Backend:   methods[0],
		Namespace: methods[1],
		Operation: methods[2],
	}
	return &method, nil
}
func (m *Transport) api() {
	m.engine.POST(m.basepath+"/api", func(c *gin.Context) {
		request := new(transport.Request)
		if err := c.ShouldBindJSON(request); err != nil {
			c.SecureJSON(200, transport.Error(transport.ReplyCodeFailure, err.Error()))
			return
		}
		if m.Security != nil {
			if !m.Security.SignVerify(request) {
				c.SecureJSON(200,
					transport.Error(transport.ReplyCodeSignInvalid, "invalid sign"))
				return
			}
		}
		if m.checkInSecure(c, request) {
			return
		}

		method, err := m.getMethod(request.Method)
		if err != nil {
			c.SecureJSON(200,
				transport.Error(transport.ReplyCodeMethodInvalid, err.Error()))
			return
		}

		req := &gological.Request{
			ID:        uuid.New().String(),
			Namespace: method.Namespace,
			Operation: method.Operation,
			Data:      []byte(request.Data),
			Headers:   c.Request.Header,
			Token:     c.Request.Header.Get(gological.AuthTokenName),
			Connection: &gological.Connection{
				RemoteAddr: GetRemoteAddr(c),
				ConnState:  c.Request.TLS,
			},
		}

		resp := m.Transport.Invoke(method.Backend, req)
		m.writerReply(c, resp)

	})
}

func (m *Transport) checkInSecure(c *gin.Context, request *transport.Request) bool {
	if m.Security != nil {

		client := &transport.Client{
			RemoteAddr: GetRemoteAddr(c),
			Referer:    c.Request.Referer(),
			UserAgent:  c.Request.UserAgent(),
		}
		if err := m.Security.Blocker(client); err != nil {
			c.SecureJSON(200,
				transport.Error(transport.ReplyCodeReqBlocked, err.Error()))
			return true
		}

		if err := m.Security.RateLimiter(client); err != nil {
			c.SecureJSON(200, transport.Error(transport.ReplyCodeRateLimited, err.Error()))
			return true
		}
	}
	return false
}

func GetRemoteAddr(c *gin.Context) string {
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr == "127.0.0.1" {
		remoteAddr = c.GetHeader("X-Forwarded-For")
	}
	return remoteAddr
}

// endpoints
func (m *Transport) endpoints() {
	routers := make([]map[string]string, 0)
	m.engine.GET(m.basepath+"/endpoints", func(c *gin.Context) {
		for _, info := range m.engine.Routes() {
			router := map[string]string{
				"path":   info.Path,
				"method": info.Method,
			}
			routers = append(routers, router)
		}
		c.Header("content-type", "application/json")
		c.Writer.WriteString(jsonutil.EncodeToString(routers))
	})
}
