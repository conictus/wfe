package wfe

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestValidateTaskFnOk(t *testing.T) {
	a := func(c *Context) {

	}

	b := func(c *Context, a int, b string, x float64) {

	}

	c := func(c *Context, a int) int {
		return a
	}

	d := func(c *Context, b string, a ...int) {

	}

	for _, fn := range []interface{}{a, b, c, d} {
		v := reflect.ValueOf(fn)

		err := validateWorkFunc(v)
		if ok := assert.Nil(t, err); !ok {
			t.Fatal()
		}
	}
}

func TestValidateTaskFnError(t *testing.T) {
	a := func() {

	}

	c := func(c *Context, a int) (int, int) {
		return a, a
	}

	for _, fn := range []interface{}{a, c} {
		v := reflect.ValueOf(fn)

		err := validateWorkFunc(v)
		if ok := assert.Error(t, err); !ok {
			t.Fatal()
		}
	}
}

func TestRegisterOK(t *testing.T) {
	a := func(c *Context) {

	}

	b := func(c *Context, a int, b string, x float64) {

	}

	c := func(c *Context, a int) int {
		return a
	}

	d := func(c *Context, b string, a ...int) {

	}

	defer func() {
		err := recover()
		if ok := assert.Nil(t, err); !ok {
			t.Fatal()
		}
	}()
	for _, fn := range []interface{}{a, b, c, d} {
		Register(fn)
	}
}

func TestRegisterPanic(t *testing.T) {
	a := func() {

	}

	c := func(c *Context, a int) (int, int) {
		return a, a
	}

	defer func() {
		err := recover()
		if ok := assert.NotNil(t, err); !ok {
			t.Fatal()
		}
	}()
	for _, fn := range []interface{}{a, c} {
		Register(fn)
	}
}
