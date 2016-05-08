package wfe

const (
	workQueue = "wfe.work"
)

var (
	//WorkQueueRoute default route options for work queue. It declares the 'wfe.work' as durable queue
	WorkQueueRoute = &RouteOptions{
		Queue:   workQueue,
		Durable: true,
	}
)

//RouteOptions specifies route and queuing options for both consumers and dispatchers.
type RouteOptions struct {
	//Queue name
	Queue string
	//Durable flag
	Durable bool
	//AutoDelete flag
	AutoDelete bool
	//Exclusive flag
	Exclusive bool
	//AutoConfirm flag (for brokers that implements delivery confirmation
	AutoConfirm bool
}

//Broker interface
type Broker interface {
	//Close broker connections. Close must force the broker Consumer to return
	Close() error

	//Dispatcher gets a dispatcher instance according to the RouterOptions specified.
	Dispatcher(o *RouteOptions) (Dispatcher, error)

	//Consumer gets a consumer instance according to the RouterOptions specified.
	Consumer(o *RouteOptions) (Consumer, error)
}

//Message content
type Message struct {
	Queue   string
	Content interface{}
}

//Dispatcher interface
type Dispatcher interface {
	//Close dispatcher
	Close() error

	//Dispatch msg
	Dispatch(msg *Message) (string, error)
}

//Consumer interfaec
type Consumer interface {
	//Close consumer
	Close() error

	//Consume gets a Deliver channel. the delivery channel is auto closed if the broker connection is lost
	//or the consumer is closed explicitly.
	Consume() (<-chan Delivery, error)
}

//Delivery interface
type Delivery interface {
	//Delivery ID
	ID() string

	//Confirm, confirms the delivery. (If needed by the broker) or NOOP if not
	Confirm() error

	/*
		Content, loads the message content

			var o Object
			if err := d.Content(&o); err != nil {
				//handle error.
			}
	*/
	Content(c interface{}) error
}
