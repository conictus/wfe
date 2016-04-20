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
