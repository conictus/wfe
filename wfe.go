/*
Package wfe is an asynchronous task queue/job queue based on distributed message passing. It is focused on real-time operation
The execution units, called tasks, are executed concurrently on a single or more worker servers using multiprocessing.
Tasks can execute asynchronously (in the background) or synchronously (wait until ready).
*/
package wfe

import (
	"context"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"runtime/debug"
	"sync"
	"time"
)

var (
	log = logging.MustGetLogger("wfe")

	//ErrUnknownFunction returned by the engine if a client is calling an unregistered function
	ErrUnknownFunction = errors.New("unkonwn function")

	DefaultQueue = Queue{DefaultQueueName, 1000}
)

//Engine is responsible for running the tasks concurrently. It processes users messages and executes them
type Engine struct {
	opt    *Options
	store  ResultStore
	graph  GraphBackend
	queues []Queue

	mw         middlewareStack
	dispatcher Dispatcher
}

type Queue struct {
	Name    string
	Workers int
}

/*
New creates a new engine with the given options and the number of workers routines. The number of workers routines
controllers how many parallel tasks can be run concurrently on this engine instance.
*/
func New(o *Options, queues ...Queue) (*Engine, error) {
	store, err := o.GetStore()
	if err != nil {
		return nil, err
	}

	graph, err := o.GetGraphBackend()
	if err != nil {
		return nil, err
	}

	return &Engine{
		opt:    o,
		store:  store,
		graph:  graph,
		queues: queues,
	}, nil
}

func (e *Engine) newContext(id string, req Request) *Context {
	return &Context{
		client: &clientImpl{
			dispatcher: e.dispatcher,
			store:      e.store,
			parentID:   id,
		},
		id:     id,
		values: make(map[string]interface{}),
	}
}

func (e *Engine) handle(id string, req Request) (interface{}, error) {
	ctx := e.newContext(id, req)
	e.mw.Enter(ctx)
	defer e.mw.Exit(ctx)

	return req.Invoke(ctx)
}

func (e *Engine) handleDelivery(delivery Delivery) error {
	defer func() {
		//we discard the message anyway
		if err := delivery.Confirm(); err != nil {
			log.Errorf("Failed to acknowledge message processing %s", err)
		}
	}()

	response := &Response{
		UUID:  delivery.ID(),
		State: StateError,
	}

	var graph Graph
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()

			log.Errorf("Message '%s' paniced: %s", delivery.ID(), err)
			response.State = StateError
			response.Error = fmt.Sprintf("%v", err)
		}

		if err := e.store.Set(response); err != nil {
			log.Errorf("Failed to send response for id (%s): %s", response.UUID, err)
		}

		if graph != nil {
			graph.Commit(response)
		}
	}()

	var req requestImpl
	if err := delivery.Content(&req); err != nil {
		response.Error = err.Error()
		return err
	}

	if e.graph != nil {
		graph, _ = e.graph.Graph(delivery.ID(), &req)
	}

	result, err := e.handle(delivery.ID(), &req)
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

func (e *Engine) startWorkers(number int) chan<- Delivery {
	ch := make(chan Delivery)
	for i := 0; i < number; i++ {
		go e.worker(ch)
	}

	return ch
}

func (e *Engine) getRequestsQueue(broker Broker, queue Queue) (<-chan Delivery, error) {
	consumer, err := broker.Consumer(&RouteOptions{
		Queue:   queue.Name,
		Durable: true,
	})

	if err != nil {
		return nil, err
	}

	requests, err := consumer.Consume()
	if err != nil {
		broker.Close()
		return nil, err
	}

	return requests, nil
}

//Use a middleware
func (e *Engine) Use(m Middleware) {
	e.mw = append(e.mw, m)
}

func (e *Engine) runQueue(ctx context.Context, wg *sync.WaitGroup, broker Broker, queue Queue) {
	requests, err := e.getRequestsQueue(broker, queue)
	if err != nil {
		log.Errorf("Faile to get queue '%v' delivereis", queue)
	}

	feed := e.startWorkers(queue.Workers)
	defer close(feed)
	for request := range requests {
		select {
		case feed <- request:
		case <-ctx.Done():
			log.Infof("Queue %s canceled", queue)
			return
		}
	}

	wg.Done()
}

//Run start processing messages.
func (e *Engine) Run() {
	for {
		broker, err := e.opt.GetBroker()
		if err != nil {
			log.Errorf("Failed to connect to broker '%s': %s", e.opt.Broker, err)
			time.Sleep(3 * time.Second)
			continue
		}

		dispatcher, err := broker.Dispatcher()

		if err != nil {
			log.Errorf("Failed to get client dispatcher")
			continue
		}

		e.dispatcher = dispatcher

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}
		wg.Add(1) //WE ALWAYS ADD ONLY ONE. SO IF ANY DIED, WE CANCEL ALL REMAINING WORKERS.

		for _, queue := range e.queues {
			go e.runQueue(ctx, &wg, broker, queue)
		}

		wg.Wait()
		cancel()
	}
}
