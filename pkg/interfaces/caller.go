package interfaces

// FunctionCallReq 函数调用请求
type FunctionCallReq struct {
	Name   string              `json:"name"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}

// FunctionCallResp 函数调用响应
type FunctionCallResp struct {
	Name   string              `json:"name"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}

// Caller 调用者接口
type Caller interface {
	Call(req *FunctionCallReq) (rsp *FunctionCallResp, err error)
}
