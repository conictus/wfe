package wfe

import (
	"encoding/gob"
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

func init() {
	gob.Register(requestImpl{})
}

type Response struct {
	UUID   string
	State  string
	Error  string
	Result interface{}
}

type Request interface {
	ID() string
	ParentID() string
	Fn() string
	Args() []interface{}
}

type PartialRequest interface {
	Request
	Append(arg interface{})
	Request() (Request, error)
}

type requestImpl struct {
	ParentUUID string
	UUID       string
	Function   string
	Arguments  []interface{}
}

func (r *requestImpl) ID() string {
	return r.UUID
}

func (r *requestImpl) ParentID() string {
	return r.ParentUUID
}

func (r *requestImpl) Fn() string {
	return r.Function
}

func (r *requestImpl) Args() []interface{} {
	return r.Arguments
}

func (r *requestImpl) String() string {
	return fmt.Sprintf("%s.%s %s(%v)", r.ParentUUID, r.UUID, r.Function, r.Arguments)
}

func (r *requestImpl) Append(arg interface{}) {
	r.Arguments = append(r.Arguments, arg)
}

func (r *requestImpl) Request() (Request, error) {
	fn, ok := fns[r.Function]
	if !ok {
		return nil, fmt.Errorf("unknown function '%s'", r.Function)
	}

	return Call(fn, r.Arguments...)
}

func expectedAt(fn reflect.Type, i int) reflect.Type {
	if fn.IsVariadic() && i >= fn.NumIn()-1 {
		argvType := fn.In(fn.NumIn() - 1)
		return argvType.Elem()
	}
	return fn.In(i)
}

func validateArgs(fn reflect.Type, partial bool, args ...interface{}) error {
	numIn := len(args)
	expectedIn := fn.NumIn() - 1 //we ignore the context arg
	if fn.IsVariadic() {
		expectedIn--
	}

	if !partial && numIn < expectedIn {
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

func makeCall(work interface{}, partial bool, args ...interface{}) (*requestImpl, error) {
	fn := reflect.ValueOf(work)
	if err := validateWorkFunc(fn); err != nil {
		return nil, err
	}

	//validate arguments list types
	t := fn.Type()
	if err := validateArgs(t, partial, args...); err != nil {
		return nil, err
	}

	call := &requestImpl{
		UUID:      uuid.New(),
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	return call, nil
}

func Call(work interface{}, args ...interface{}) (Request, error) {
	return makeCall(work, false, args...)
}

func PartialCall(work interface{}, args ...interface{}) (PartialRequest, error) {
	return makeCall(work, true, args...)
}

func MustCall(work interface{}, args ...interface{}) Request {
	req, err := Call(work, args...)
	if err != nil {
		panic(err)
	}
	return req
}

func MustPartialCall(work interface{}, args ...interface{}) PartialRequest {
	req, err := PartialCall(work, args...)
	if err != nil {
		panic(err)
	}
	return req
}
