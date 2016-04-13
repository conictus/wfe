package wfe

type Broker interface {
	//client side
	Call(call Call) error
	Respond(queue string, response Response) error

	Responses(queue string) (<-chan Response, error)
	//server side
	Requests() (<-chan Request, error)

	Close()
}
