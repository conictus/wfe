package wfe

type Request interface {
	Ack() error
	Call() (Call, error)
	Respond(Response) error
}
