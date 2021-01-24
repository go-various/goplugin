package http

import (
	"github.com/gin-gonic/gin"
	"github.com/go-various/goplugin"
	gological "github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/goplugin/transport/logical"
)

func (m *Transport) writerReply(c *gin.Context, resp *workerReply) {

	if m.logger.IsTrace() {
		m.traceReply(c, resp)
	}

	if resp.err == pluginregister.PluginNotExists {
		reply := logical.Error(
			logical.ReplyCodeBackendNotExists,
			resp.err.Error())

		c.SecureJSON(200, reply)
		return
	}

	if resp.err == goplugin.ErrPluginShutdown {
		reply := logical.Error(
			logical.ReplyCodeBackendShutdown,
			resp.err.Error())

		c.SecureJSON(200, reply)
		return
	}
	if resp.err == gological.ErrAuthorizationTokenRequired || resp.err == gological.ErrAuthorizationTokenInvalid {
		reply := logical.Error(logical.ReplyCodeAuthorizedRequired, resp.err.Error())
		c.SecureJSON(200, reply)
		return
	}
	if nil != resp.err {
		reply := logical.Error(logical.ReplyCodeFailure, resp.err.Error())
		c.SecureJSON(200, reply)
		return
	}
	c.SecureJSON(200, logical.Success(resp.result))
}

func (m *Transport) traceReply(c *gin.Context, resp *workerReply) {
	if m.logger.IsTrace() {
		m.logger.Trace(
			"http-gateway trace",
			"path", c.Request.RequestURI,
			"method", c.Request.Method,
			"client", GetRemoteAddr(c),
			"err", resp.err,
		)
	}
}
