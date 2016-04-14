package wfe

//
//type Broker interface {
//	//client side
//	Call(call Call) error
//	Respond(queue string, response Response) error
//
//	Responses(queue string) (<-chan Response, error)
//	//server side
//	Requests() (<-chan Request, error)
//
//	Close()
//}

const (
	WorkQueue = "wfe.work"
)

var (
	WorkQueueRoute = &RouteOptions{
		Queue:   WorkQueue,
		Durable: true,
	}
)

type RouteOptions struct {
	Queue       string
	Durable     bool
	AutoDelete  bool
	Exclusive   bool
	AutoConfirm bool
}

type Broker interface {
	Close() error
	Dispatcher(o *RouteOptions) (Dispatcher, error)
	Consumer(o *RouteOptions) (Consumer, error)
}

type Message struct {
	ID      string
	Queue   string
	Content interface{}
	ReplyTo string
}

type Dispatcher interface {
	Close() error
	Dispatch(msg *Message) error
}

type Consumer interface {
	Close() error
	Consume() (<-chan Delivery, error)
}

type Delivery interface {
	ID() string
	Confirm() error
	ReplyQueue() string
	Content(c interface{}) error
}
