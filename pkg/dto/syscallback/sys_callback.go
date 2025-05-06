package syscallback

type Response struct {
	Data interface{} `json:"data"`
}

type ResponseWith[T any] struct {
	Data T `json:"data"`
}
