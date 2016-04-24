package wfe

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

var (
	fns = make(map[string]interface{})
	m   sync.Mutex
)

func validateWorkFunc(v reflect.Value) error {
	if v.Kind() != reflect.Func {
		return fmt.Errorf("not a function")
	}
	t := v.Type()
	if t.NumIn() == 0 {
		return fmt.Errorf("worker function must accept at least one argument of type *wfe.Context")
	}

	if t.In(0) != reflect.TypeOf((*Context)(nil)) {
		return fmt.Errorf("worker function first argument not of type *wfe.Context")
	}

	if t.NumOut() > 1 {
		return fmt.Errorf("worker function must return maximum of one object")
	}

	return nil
}

/*
Register a task function. A task function must accept a `*Context` as first argument followed by and any number of arguments
needed by the task. A task function can return zero or one object.
Register panics if the task signature is wrongThe register process usually happens inside an init function

Example:
	func Add(c *gin.Context, args ...int) {
		v := 0
		for i := 0; i < len(args); i++ {
			v += i
		}
		return v
	}

	func init() {
		wfe.Register(Add)
	}

Note: A task can return an error by a panic
*/
func Register(fn interface{}) {
	v := reflect.ValueOf(fn)
	if err := validateWorkFunc(v); err != nil {
		panic(err)
	}

	n := runtime.FuncForPC(v.Pointer()).Name()
	log.Debugf("Registering function '%s'", n)
	m.Lock()
	defer m.Unlock()
	fns[n] = fn
}
