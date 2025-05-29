package runcher

type RunFunctionReq struct {
	User     string `json:"user"`
	Runner   string `json:"runner"`
	Method   string `json:"method"`
	Router   string `json:"router"`
	Body     string `json:"body"`
	Version  string `json:"version"`
	RawQuery string `json:"raw_query"`
}
