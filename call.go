package wfe

const (
	StateSuccess = "success"
	StateError   = "error"
)

type Call struct {
	UUID      string
	Function  string
	Arguments []interface{}
}

type Response struct {
	State   string
	Error   string
	Results []interface{}
}
