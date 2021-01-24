package transport

import "github.com/go-various/goplugin/logical"

// SchemaResponse
// 客户端请求schema的返回结构体
type SchemaResponse struct {
	Name       string             `json:"name"`
	Backend    string             `json:"backend"`
	Namespaces logical.Namespaces `json:"namespaces"`
}

type Client struct {
	RemoteAddr string
	Referer    string
	UserAgent  string
}

type Method struct {
	Backend   string
	Namespace string
	Operation string
}

type Request struct {
	Method    string `json:"method" binding:"required"` //${backend}.${namespace}.${operation}
	Version   string `json:"version" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
	SignType  string `json:"sign_type" binding:"required"`
	Sign      string `json:"sign" binding:"required"`
	Data      string `json:"data" binding:"required"`
}

// http返回数据结构
type Response struct {
	Code    ReplyCode   `json:"code"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
}

func Error(code ReplyCode, message string) *Response {
	return &Response{Code: code, Message: message}
}

func Success(result interface{}) *Response {
	return &Response{Code: 0, Result: result}
}

type Transport interface {
	SetSecurity(security Security)
	SetAuthMethod(method string) error
	SetAuthEnabled()
	SetAuthDisabled()
	Shutdown()
	Running() <-chan bool
	Listen(addr string, port uint) error
	Start() error

	// AddHandle
	// in http: gin.HandlerFunc
	// in http: gin.HandlerFunc, httpMethod, relativePath
	// in rpc : RPC Service instance, eg &Service{}
	AddHandle(handle interface{}, args ...string) error
}

type Security interface {
	SignVerify(args *Request) bool
	RateLimiter(client *Client) error
	Blocker(client *Client) error
}
