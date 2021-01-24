package logical

// http返回数据结构
type Reply struct {
	Code    ReplyCode `json:"code"`
	Result  interface{}      `json:"result,omitempty"`
	Message string           `json:"message,omitempty"`
}

func Error(code ReplyCode, message string) *Reply {
	return &Reply{Code: code, Message: message}
}

func Success(result interface{}) *Reply {
	return &Reply{Code: 0, Result: result}
}
