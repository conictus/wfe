package wfe

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"reflect"
	"time"
)

var (
	log             = logging.MustGetLogger("wfe")
	UnknownFunction = errors.New("unkonwn function")
)

type Engine struct {
	opt     *Options
	store   ResultStore
	workers int

	client Client
}

func New(o *Options, workers int) (*Engine, error) {
	store, err := o.GetStore()
	if err != nil {
		log.Warningf("Failed to create a result store: %s", err)
	}

	return &Engine{
		opt:     o,
		store:   store,
		workers: workers,
	}, nil
}

func (e *Engine) newContext(req Request) *Context {
	return &Context{
		Client: e.client,
		id:     req.ID(),
	}
}

func (e *Engine) handle(req Request) (ResultTuple, error) {
	log.Debugf("Calling %s", req)
	fn, ok := fns[req.Fn()]
	if !ok {
		return nil, UnknownFunction
	}

	callable := reflect.ValueOf(fn)
	callableType := callable.Type()

	values := make([]reflect.Value, 0)
	values = append(values, reflect.ValueOf(e.newContext(req)))

	for i, arg := range req.Args() {
		argType := expectedAt(callableType, i+1)
		inValue := reflect.ValueOf(arg)

		switch argType.Kind() {
		case reflect.Ptr:
			fallthrough
		case reflect.Interface:
			new := reflect.New(inValue.Type())
			new.Elem().Set(inValue)
			inValue = new
		}

		values = append(values, inValue)
	}

	returns := callable.Call(values)

	results := make(ResultTuple, 0, len(values))
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
		if err := recover(); err != nil {
			log.Errorf("Message '%s' paniced: %s", delivery.ID(), err)
			response.State = StateError
			switch x := err.(type) {
			case error:
				response.Error = x.Error()
			case fmt.Stringer:
				response.Error = x.String()
			case string:
				response.Error = x
			default:
				response.Error = "unknown error"
			}
		}

		if err := e.store.Set(&response); err != nil {
			log.Errorf("Failed to send response for id (%s): %s", response.UUID, err)
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
}

func (e *Engine) startWorkers() chan<- Delivery {
	ch := make(chan Delivery)
	for i := 0; i < e.workers; i++ {
		go e.worker(ch)
	}

	return ch
}

func (e *Engine) getRequestsQueue() (Broker, <-chan Delivery, error) {
	broker, err := e.opt.GetBroker()
	if err != nil {
		return nil, nil, err
	}

	consumer, err := broker.Consumer(WorkQueueRoute)
	if err != nil {
		return nil, nil, err
	}

	requests, err := consumer.Consume()
	if err != nil {
		broker.Close()
		return nil, nil, err
	}

	return broker, requests, nil
}

func (e *Engine) getClient(broker Broker) (Client, error) {
	dispatcher, err := broker.Dispatcher(WorkQueueRoute)

	if err != nil {
		return nil, err
	}

	return &clientImpl{
		dispatcher: dispatcher,
		store:      e.store,
	}, nil
}

func (e *Engine) Run() {
	for {
		broker, requests, err := e.getRequestsQueue()
		if err != nil {
			log.Errorf("Failed to connect to broker '%s': %s", e.opt.Broker, err)
			time.Sleep(3 * time.Second)
			continue
		}

		client, err := e.getClient(broker)
		if err != nil {
			log.Errorf("Failed to intialize client: %s", err)
		}

		e.client = client

		feed := e.startWorkers()
		for request := range requests {
			feed <- request
		}
		log.Warningf("Lost connection with broker")

		broker.Close()
		client.Close()

		close(feed)
	}
}
