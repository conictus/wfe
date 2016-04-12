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
	Call(work interface{}, args ...interface{}) (Job, error)
}

type clientIml struct {
	broker Broker
}

func NewClient(broker Broker) Client {
	return &clientIml{
		broker: broker,
	}
}

func (c *clientIml) expectedAt(fn reflect.Type, i int) reflect.Type {
	if fn.IsVariadic() && i >= fn.NumIn()-1 {
		argvType := fn.In(fn.NumIn() - 1)
		return argvType.Elem()
	}
	return fn.In(i)
}

func (c *clientIml) Call(work interface{}, args ...interface{}) (Job, error) {
	fn := reflect.ValueOf(work)
	if err := validateWorkFunc(fn); err != nil {
		return nil, err
	}

	//validate arguments list types
	t := fn.Type()

	numIn := len(args)
	expectedIn := t.NumIn() - 1 //we ignore the context arg
	if t.IsVariadic() {
		expectedIn--
	}

	if numIn < expectedIn {
		return nil, TooFewArgumentsError
	}
	if !t.IsVariadic() && numIn > expectedIn {
		return nil, TooManyArgumentsError
	}

	for i, arg := range args {
		actual := reflect.TypeOf(arg)
		expected := c.expectedAt(t, i+1)

		if !actual.AssignableTo(expected) {
			return nil, fmt.Errorf("argument type mismatch at position %d expected %s", i+1, expected)
		}
	}

	call := Call{
		ID:        uuid.New(),
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	if err := c.broker.Dispatch(call); err != nil {
		return nil, err
	}
	return &call, nil
}
