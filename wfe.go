/*
Package wfe is an asynchronous task queue/job queue based on distributed message passing. It is focused on real-time operation
The execution units, called tasks, are executed concurrently on a single or more worker servers using multiprocessing.
Tasks can execute asynchronously (in the background) or synchronously (wait until ready).
*/
package wfe

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"reflect"
	"time"
)

var (
	log = logging.MustGetLogger("wfe")

	//ErrUnknownFunction returned by the engine if a client is calling an unregistered function
	ErrUnknownFunction = errors.New("unkonwn function")
)

//Engine is responsible for running the tasks concurrently. It processes users messages and executes them
type Engine struct {
	opt     *Options
	store   ResultStore
	workers int

	mw         middlewareStack
	dispatcher Dispatcher
}

/*
New creates a new engine with the given options and the number of workers routines. The number of workers routines
controllers how many parallel tasks can be run concurrently on this engine instance.
*/
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
		client: &clientImpl{
			dispatcher: e.dispatcher,
			store:      e.store,
			parentID:   req.ID(),
		},
		id:     req.ID(),
		values: make(map[string]interface{}),
	}
}

func (e *Engine) handle(req Request) (interface{}, error) {
	log.Debugf("Calling %s", req)
	fn, ok := fns[req.Fn()]
	if !ok {
		return nil, ErrUnknownFunction
	}

	callable := reflect.ValueOf(fn)
	callableType := callable.Type()

	var values []reflect.Value
	ctx := e.newContext(req)
	e.mw.Enter(ctx)
	defer e.mw.Exit(ctx)

	values = append(values, reflect.ValueOf(ctx))

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

	var result interface{}
	if len(returns) == 1 {
		result = returns[0].Interface()
	}

	return result, nil
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
			response.Error = fmt.Sprintf("%v", err)
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

	result, err := e.handle(&req)
	if err != nil {
		response.Error = err.Error()
		return err
	}

	response.State = StateSuccess
	response.Result = result

	return nil
}

func (e *Engine) worker(queue <-chan Delivery) {
	for request := range queue {
		log.Debugf("received message: %s", request.ID())
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

//Use a middleware
func (e *Engine) Use(m Middleware) {
	e.mw = append(e.mw, m)
}

//Run start processing messages.
func (e *Engine) Run() {
	for {
		broker, requests, err := e.getRequestsQueue()
		if err != nil {
			log.Errorf("Failed to connect to broker '%s': %s", e.opt.Broker, err)
			time.Sleep(3 * time.Second)
			continue
		}

		dispatcher, err := broker.Dispatcher(WorkQueueRoute)
		if err != nil {
			log.Errorf("Failed to intialize client: %s", err)
		}

		e.dispatcher = dispatcher

		feed := e.startWorkers()
		for request := range requests {
			feed <- request
		}
		log.Warningf("Lost connection with broker")

		broker.Close()
		dispatcher.Close()

		close(feed)
	}
}
