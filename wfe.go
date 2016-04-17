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
	store   ResultStore
	workers int

	client Client
}

func New(o *Options, workers int) (*Engine, error) {
	broker, err := o.GetBroker()
	if err != nil {
		return nil, err
	}

	store, err := o.GetStore()
	if err != nil {
		return nil, err
	}

	client, err := NewClient(o)

	if err != nil {
		return nil, err
	}
	return &Engine{
		broker:  broker,
		store:   store,
		workers: workers,
		client:  client,
	}, nil
}

func (e *Engine) newContext(req Request) *Context {
	return &Context{
		Client: e.client,
		id:     req.ID(),
	}
}

func (e *Engine) handle(req Request) ([]interface{}, error) {
	fn, ok := fns[req.Fn()]
	if !ok {
		return nil, UnknownFunction
	}

	values := make([]reflect.Value, 0)
	values = append(values, reflect.ValueOf(e.newContext(req)))

	for _, arg := range req.Args() {
		values = append(values, reflect.ValueOf(arg))
	}

	callable := reflect.ValueOf(fn)
	returns := callable.Call(values)

	results := make([]interface{}, 0, len(values))
	for _, value := range returns {
		var v interface{}
		switch x := value.Interface().(type) {
		case error:
			v = Error{x.Error()}
		default:
			v = x
		}

		results = append(results, v)
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

	response := Response{
		UUID:  delivery.ID(),
		State: StateError,
	}

	defer func() {
		if err := e.store.Set(&response); err != nil {
			log.Errorf("Failed to send response for id: %s", response.UUID)
		}
	}()

	var req requestImpl
	if err := delivery.Content(&req); err != nil {
		response.Error = err.Error()
		return err
	}

	results, err := e.handle(&req)
	if err != nil {
		response.Error = err.Error()
		return err
	}

	response.State = StateSuccess
	response.Results = results

	return nil
}

func (e *Engine) worker(queue <-chan Delivery) {
	for request := range queue {
		log.Debugf("received request: %s", request.ID())
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
