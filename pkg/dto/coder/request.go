package coder

type AddApiReq struct {
	Runner  *Runner  `json:"runner"`
	CodeApi *CodeApi `json:"code_api"`
}

type AddApisReq struct {
	Runner   *Runner
	CodeApis []*CodeApi
}
