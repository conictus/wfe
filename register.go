package wfe

import (
	"fmt"
	"reflect"
	"runtime"
)

var (
	registered map[string]interface{}
)

func init() {
	registered = make(map[string]interface{})
}

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

	return nil
}

func Register(fn interface{}) {
	v := reflect.ValueOf(fn)
	if err := validateWorkFunc(v); err != nil {
		panic(err)
	}

	n := runtime.FuncForPC(v.Pointer()).Name()
	registered[n] = fn
}
