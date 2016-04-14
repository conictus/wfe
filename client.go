package wfe

import (
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"reflect"
	"runtime"
)

var (
	TooFewArgumentsError  = errors.New("call with too few arguments")
	TooManyArgumentsError = errors.New("call with too many arguments")
)

type Client interface {
	Call(work interface{}, args ...interface{}) (Result, error)
}

type clientImpl struct {
	dispatcher Dispatcher
	consumer   Consumer
	replyTo    string
}

func NewClient(broker Broker) (Client, error) {
	dispatcher, err := broker.Dispatcher(WorkQueueRoute)

	if err != nil {
		return nil, err
	}
	replyTo := fmt.Sprintf("client.%s", uuid.New())
	consumer, err := broker.Consumer(&RouteOptions{
		Queue:       replyTo,
		Exclusive:   true,
		AutoConfirm: true,
		AutoDelete:  true,
	})

	if err != nil {
		return nil, err
	}

	c := &clientImpl{
		replyTo:    replyTo,
		dispatcher: dispatcher,
		consumer:   consumer,
	}

	c.receiveResponses()
	return c, nil
}

func (c *clientImpl) Close() {
	c.consumer.Close()
	c.dispatcher.Close()
}

func (c *clientImpl) receiveResponses() {
	go func() {
		deliveries, err := c.consumer.Consume()
		if err != nil {
			log.Errorf("Failed to receive responses:", err)
			return
		}

		for delivery := range deliveries {
			var results []interface{}
			delivery.Content(&results)
			log.Infof("Got response for %s", results)
		}
	}()
}

func (c *clientImpl) expectedAt(fn reflect.Type, i int) reflect.Type {
	if fn.IsVariadic() && i >= fn.NumIn()-1 {
		argvType := fn.In(fn.NumIn() - 1)
		return argvType.Elem()
	}
	return fn.In(i)
}

func (c *clientImpl) validateArgs(fn reflect.Type, args ...interface{}) error {
	numIn := len(args)
	expectedIn := fn.NumIn() - 1 //we ignore the context arg
	if fn.IsVariadic() {
		expectedIn--
	}

	if numIn < expectedIn {
		return TooFewArgumentsError
	}
	if !fn.IsVariadic() && numIn > expectedIn {
		return TooManyArgumentsError
	}

	for i, arg := range args {
		actual := reflect.TypeOf(arg)
		expected := c.expectedAt(fn, i+1)

		if !actual.AssignableTo(expected) {
			return fmt.Errorf("argument type mismatch at position %d expected %s", i+1, expected)
		}
	}

	return nil
}

func (c *clientImpl) Call(work interface{}, args ...interface{}) (Result, error) {
	fn := reflect.ValueOf(work)
	if err := validateWorkFunc(fn); err != nil {
		return nil, err
	}

	//validate arguments list types
	t := fn.Type()
	if err := c.validateArgs(t, args...); err != nil {
		return nil, err
	}

	call := Call{
		UUID:      uuid.New(),
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	msg := Message{
		ID:      call.UUID,
		ReplyTo: c.replyTo,
		Content: call,
	}

	if err := c.dispatcher.Dispatch(&msg); err != nil {
		return nil, err
	}

	return &call, nil
}
