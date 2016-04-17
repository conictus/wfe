package wfe

import (
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"reflect"
	"runtime"
)

const (
	StateSuccess = "success"
	StateError   = "error"
)

var (
	ErrTooFewArguments  = errors.New("call with too few arguments")
	ErrTooManyArguments = errors.New("call with too many arguments")
)

type Request interface {
	ID() string
	Fn() string
	Args() []interface{}
}

type requestImpl struct {
	UUID      string
	Function  string
	Arguments []interface{}
}

type Response struct {
	UUID    string
	State   string
	Error   string
	Results []interface{}
}

func (r *requestImpl) ID() string {
	return r.UUID
}

func (r *requestImpl) Fn() string {
	return r.Function
}

func (r *requestImpl) Args() []interface{} {
	return r.Arguments
}

func expectedAt(fn reflect.Type, i int) reflect.Type {
	if fn.IsVariadic() && i >= fn.NumIn()-1 {
		argvType := fn.In(fn.NumIn() - 1)
		return argvType.Elem()
	}
	return fn.In(i)
}

func validateArgs(fn reflect.Type, args ...interface{}) error {
	numIn := len(args)
	expectedIn := fn.NumIn() - 1 //we ignore the context arg
	if fn.IsVariadic() {
		expectedIn--
	}

	if numIn < expectedIn {
		return ErrTooFewArguments
	}
	if !fn.IsVariadic() && numIn > expectedIn {
		return ErrTooManyArguments
	}

	for i, arg := range args {
		actual := reflect.TypeOf(arg)
		expected := expectedAt(fn, i+1)

		if !actual.AssignableTo(expected) {
			return fmt.Errorf("argument type mismatch at position %d expected %s", i+1, expected)
		}
	}

	return nil
}

func Call(work interface{}, args ...interface{}) (Request, error) {
	fn := reflect.ValueOf(work)
	if err := validateWorkFunc(fn); err != nil {
		return nil, err
	}

	//validate arguments list types
	t := fn.Type()
	if err := validateArgs(t, args...); err != nil {
		return nil, err
	}

	call := &requestImpl{
		UUID:      uuid.New(),
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	return call, nil
}
