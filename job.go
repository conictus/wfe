package wfe

type Job interface {
	ID() string
	Get() []interface{}
}
