package wfe

import (
	"fmt"
	"github.com/op/go-logging"
	"reflect"
)

var (
	log = logging.MustGetLogger("wfe")
)

type Engine struct {
	broker  Broker
	workers int
}

func New(broker Broker, workers int) *Engine {
	return &Engine{
		broker:  broker,
		workers: workers,
	}
}

func (e *Engine) handle(request CallRequest) error {
	call, err := request.Call()
	if err != nil {
		return err
	}

	fn, ok := registered[call.Function]
	if !ok {
		return fmt.Errorf("calling unregisted function '%s'", call.Function)
	}
	values := make([]reflect.Value, 0)
	values = append(values, reflect.ValueOf(&Context{}))
	for _, arg := range call.Arguments {
		values = append(values, reflect.ValueOf(arg))
	}
	callable := reflect.ValueOf(fn)
	callable.Call(values)
	return nil
}

func (e *Engine) worker(q <-chan CallRequest) {
	for {
		request := <-q
		if err := e.handle(request); err != nil {
			log.Errorf("Failed to handle request: %s", err)
		}
		request.Ack()
	}
}

func (e *Engine) init() chan<- CallRequest {
	ch := make(chan CallRequest)
	for i := 0; i < e.workers; i++ {
		go e.worker(ch)
	}

	return ch
}

func (e *Engine) Run() error {
	requests, err := e.broker.Consume()
	if err != nil {
		return err
	}

	feed := e.init()
	for request := range requests {
		feed <- request
	}

	return nil
}
