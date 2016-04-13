package wfe

type Response interface {
	ID() string
	Get() []interface{}
}
