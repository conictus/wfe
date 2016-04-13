package wfe

const (
	WorkQueue       = "wfe.work"
	contentType     = "application/wfe+call"
	contentEncoding = "encoding/gob"
)

type Request interface {
	Ack() error
	Call() (Call, error)
}

type Broker interface {
	Dispatch(call Call) error
	Consume() (<-chan Request, error)
}
