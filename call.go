package wfe

import (
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

const (
	//StateSuccess notates tasks has exited without an error
	StateSuccess = "success"

	//StateError notates tasks has exited with an error
	StateError = "error"
)

var (
	//ErrTooFewArguments task expecting more arguments than provided.
	ErrTooFewArguments = errors.New("call with too few arguments")

	//ErrTooManyArguments task expecting fewer arguments than provided.
	ErrTooManyArguments = errors.New("call with too many arguments")
)

func init() {
	gob.Register(requestImpl{})
}

//Response returned by the wfe after a task finishes.
type Response struct {
	//UUID request UUID
	UUID string
	//Exit state of task execution (StateSuccess, StateError)
	State string
	//Error message if State != StateSuccess
	Error string
	//Result object returned by the task
	Result interface{}
}

type ParentIDSetter interface {
	SetParentID(id string)
}

//Request interface
type Request interface {
	ParentID() string
	Fn() string
	Args() []interface{}
}

//PartialRequest is a request with fewer arguments than the task expect.
//Is used to build callbacks and chains when a response of one task is fed to another in a chain
type PartialRequest interface {
	Request
	Append(arg interface{})
	Request() (Request, error)
	MustRequest() Request
}

type requestImpl struct {
	ParentUUID string
	Function   string
	Arguments  []interface{}
}

func (r *requestImpl) ParentID() string {
	return r.ParentUUID
}

func (r *requestImpl) SetParentID(id string) {
	r.ParentUUID = id
}

func (r *requestImpl) Fn() string {
	return r.Function
}

func (r *requestImpl) Args() []interface{} {
	return r.Arguments
}

func (r *requestImpl) String() string {
	return fmt.Sprintf("parent(%s) %s(%v)", r.ParentUUID, r.Function, r.Arguments)
}

func (r *requestImpl) Append(arg interface{}) {
	r.Arguments = append(r.Arguments, arg)
}

func (r *requestImpl) Request() (Request, error) {
	fn, ok := Registered(r.Function)
	if !ok {
		return nil, fmt.Errorf("unknown function '%s'", r.Function)
	}

	req, err := Call(fn, r.Arguments...)
	if err != nil {
		return nil, err
	}

	if r.ParentID() != "" {
		if req, ok := req.(ParentIDSetter); ok {
			req.SetParentID(r.ParentUUID)
		}
	}

	return req, nil
}

func (r *requestImpl) MustRequest() Request {
	req, err := r.Request()
	if err != nil {
		panic(err)
	}

	return req
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
			return fmt.Errorf("argument type mismatch at position %d expected %s got '%s' instead", i+1, expected, actual)
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
		Function:  runtime.FuncForPC(fn.Pointer()).Name(),
		Arguments: args,
	}

	return call, nil
}

/*
Call creates a new `Request`
Example:
	func Task(c *Context, a, b int) int {
		return a + b
	}

	//from the caller
	req, _ := wfe.Call(Task, 1, 2)
	client.Apply(req)
A call will fail to create a request if the number of arguments doesn't match the required arguments of the task function
or if the types don't match.
*/
func Call(work interface{}, args ...interface{}) (Request, error) {
	return makeCall(work, false, args...)
}

/*
PartialCall create a new `PartialCall`. Partial calls can be used in chains and chords as a callbacks.
*/
func PartialCall(work interface{}, args ...interface{}) (PartialRequest, error) {
	return makeCall(work, true, args...)
}

/*
MustCall creates a new call request, panics if the request can't be created.
*/
func MustCall(work interface{}, args ...interface{}) Request {
	req, err := Call(work, args...)
	if err != nil {
		panic(err)
	}
	return req
}

/*
MustPartialCall create a new `PartialRequest`, panics if the request can't be created
*/
func MustPartialCall(work interface{}, args ...interface{}) PartialRequest {
	req, err := PartialCall(work, args...)
	if err != nil {
		panic(err)
	}
	return req
}
