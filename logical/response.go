package logical

// Response is a struct that stores the response of a request.
// It is used to abstract the details of the higher level request protocol.
type Response struct {
	ResultCode int64               `xml:"result_code" json:"result_code" structs:"result_code" mapstructure:"result_code"`
	ResultMsg  string              `xml:"result_msg" json:"result_msg,omitempty" structs:"result_msg" mapstructure:"result_msg"`
	Content    *Content            `xml:"content" json:"result_content,omitempty" structs:"data" mapstructure:"data"`
	Headers    map[string][]string `xml:"headers" json:"headers,omitempty" structs:"headers" mapstructure:"headers"`
}

type Content struct {
	Data       interface{} `json:"data,omitempty" xml:"data"`
	Pagination interface{} `json:"pagination,omitempty" xml:"pagination"`
}
