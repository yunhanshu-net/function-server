package runcher

type Body struct {
	DataType string                 `json:"data_type"`
	TraceID  string                 `json:"trace_id"`
	MetaData map[string]interface{} `json:"meta_data"` //sdk 层
	Code     int                    `json:"code"`
	Msg      string                 `json:"msg"`
	Data     interface{}            `json:"data"`
}
