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
	Call(work interface{}, args ...interface{}) (Response, error)
}

type clientImpl struct {
	broker Broker
}

func NewClient(broker Broker) Client {
	return &clientImpl{
		broker: broker,
	}
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

func (c *clientImpl) Call(work interface{}, args ...interface{}) (Response, error) {
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
		UUID:        uuid.New(),
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	if err := c.broker.Dispatch(call); err != nil {
		return nil, err
	}

	return &call, nil
}
