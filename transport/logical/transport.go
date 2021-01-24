package logical

import "github.com/go-various/goplugin/logical"

// SchemaResponse
// 客户端请求schema的返回结构体
type SchemaResponse struct {
	Name       string     `json:"name"`
	Backend    string     `json:"backend"`
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
type RequestArgs struct {
	Method    string `json:"method" binding:"required"` //${backend}.${namespace}.${operation}
	Version   string `json:"version" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
	SignType  string `json:"sign_type" binding:"required"`
	Sign      string `json:"sign" binding:"required"`
	Data      string `json:"data" binding:"required"`
}


type Transport interface {
	SetSecurity(security Security)
	SetAuthMethod(method string) error
	SetAuthEnabled()
	SetAuthDisabled()
	Shutdown()
	Running() <-chan bool
	Listen(addr string, port uint) error
	Serve(basePath string) error
}

