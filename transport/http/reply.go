package http

import (
	"github.com/gin-gonic/gin"
	"github.com/go-various/goplugin"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport"
	"github.com/go-various/goplugin/transport/logical"
)

func (m *Transport) writerReply(c *gin.Context, resp *logical.WorkerReply) {

	if m.Logger.IsTrace() {
		m.traceReply(c, resp)
	}

	if resp.Err == pluginregister.PluginNotExists {
		reply := transport.Error(
			transport.ReplyCodeBackendNotExists,
			resp.Err.Error())
		c.SecureJSON(200, reply)
		return
	}

	if resp.Err == goplugin.ErrPluginShutdown {
		reply := transport.Error(
			transport.ReplyCodeBackendShutdown,
			resp.Err.Error())

		c.SecureJSON(200, reply)
		return
	}
	if resp.Err == gological.ErrAuthorizationTokenRequired || resp.Err == gological.ErrAuthorizationTokenInvalid {
		reply := transport.Error(transport.ReplyCodeAuthorizedRequired, resp.Err.Error())
		c.SecureJSON(200, reply)
		return
	}
	if nil != resp.Err {
		reply := transport.Error(transport.ReplyCodeFailure, resp.Err.Error())
		c.SecureJSON(200, reply)
		return
	}
	c.SecureJSON(200, transport.Success(resp.Result))
}

func (m *Transport) traceReply(c *gin.Context, resp *logical.WorkerReply) {
	if m.Logger.IsTrace() {
		m.Logger.Trace(
			"http-gateway trace",
			"path", c.Request.RequestURI,
			"method", c.Request.Method,
			"client", GetRemoteAddr(c),
			"err", resp.Err,
		)
	}
}
