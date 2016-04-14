package wfe

import (
	"errors"
	"github.com/op/go-logging"
	"reflect"
)

var (
	log             = logging.MustGetLogger("wfe")
	UnknownFunction = errors.New("unkonwn function")
)

type Engine struct {
	broker  Broker
	workers int

	dispatcher Dispatcher
}

func New(broker Broker, workers int) *Engine {
	return &Engine{
		broker:  broker,
		workers: workers,
	}
}

func (e *Engine) newContext(call *Call) *Context {
	return &Context{
		id: call.UUID,
	}
}

func (e *Engine) handle(call *Call) ([]interface{}, error) {
	fn, ok := registered[call.Function]
	if !ok {
		return nil, UnknownFunction
	}

	values := make([]reflect.Value, 0)
	values = append(values, reflect.ValueOf(e.newContext(call)))

	for _, arg := range call.Arguments {
		values = append(values, reflect.ValueOf(arg))
	}

	callable := reflect.ValueOf(fn)
	returns := callable.Call(values)

	results := make([]interface{}, 0, len(values))
	for _, value := range returns {
		results = append(results, value.Interface())
	}

	return results, nil
}

func (e *Engine) handleDelivery(delivery Delivery) error {
	defer func() {
		//we discard the message anyway
		if err := delivery.Confirm(); err != nil {
			log.Errorf("Failed to acknowledge message processing %s", err)
		}
	}()

	var call Call
	if err := delivery.Content(&call); err != nil {
		return err
	}

	results, err := e.handle(&call)
	if err != nil {
		return err
	}

	return e.dispatcher.Dispatch(&Message{
		ID:      delivery.ID(),
		Queue:   delivery.ReplyQueue(),
		Content: results,
	})
}

func (e *Engine) worker(queue <-chan Delivery) {
	for request := range queue {
		if err := e.handleDelivery(request); err != nil {
			log.Errorf("Failed to handle message: %s", err)
		}
	}

	log.Errorf("worker routine exited")
}

func (e *Engine) init() chan<- Delivery {
	ch := make(chan Delivery)
	for i := 0; i < e.workers; i++ {
		go e.worker(ch)
	}

	return ch
}

func (e *Engine) Run() error {
	dispatcher, err := e.broker.Dispatcher(nil)
	if err != nil {
		return err
	}

	e.dispatcher = dispatcher

	consumer, err := e.broker.Consumer(WorkQueueRoute)
	if err != nil {
		return err
	}
	requests, err := consumer.Consume()
	if err != nil {
		return err
	}

	feed := e.init()
	defer close(feed)

	for request := range requests {
		feed <- request
	}

	return nil
}
