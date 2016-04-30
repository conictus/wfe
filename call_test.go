package wfe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCallSuccess(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	req, err := Call(x, "test", 10)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Implements(t, (*Request)(nil), req); !ok {
		t.Fatal()
	}
}

func TestCallErrorWrongTypes(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := Call(x, 10, "test")
	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}
}

func TestCallErrorFewArgs(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := Call(x, "test")

	if ok := assert.Equal(t, ErrTooFewArguments, err); !ok {
		t.Fatal()
	}
}

func TestCallErrorManyArgs(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := Call(x, "test", 20, 30)

	if ok := assert.Equal(t, ErrTooManyArguments, err); !ok {
		t.Fatal()
	}
}

func TestCallVariadicSuccessNoVars(t *testing.T) {
	x := func(c *Context, a string, b ...int) {

	}
	_, err := Call(x, "test")
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}
}

func TestCallVariadicSuccessArgs(t *testing.T) {
	x := func(c *Context, a string, b ...int) {

	}
	_, err := Call(x, "test", 20, 30, 40)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}
}

func TestCallVariadicErrorTooFew(t *testing.T) {
	x := func(c *Context, a string, b ...int) {

	}
	_, err := Call(x)
	if ok := assert.Equal(t, ErrTooFewArguments, err); !ok {
		t.Fatal()
	}
}

func TestCallVariadicErrorWrongTypes(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := Call(x, "test", "2", "3")
	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}
}

func TestPartialCallSuccessFewArgs(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := PartialCall(x, "test")

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}
}

func TestPartialCallErrorManyArgs(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	_, err := PartialCall(x, "test", 20, 30)

	if ok := assert.Equal(t, ErrTooManyArguments, err); !ok {
		t.Fatal()
	}
}

func TestMustCallPanic(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	defer func() {
		err := recover()
		if ok := assert.NotNil(t, err); !ok {
			t.Fatal()
		}
	}()

	MustCall(x, 10)
}

func TestMustPartialCallPanic(t *testing.T) {
	x := func(c *Context, a string, b int) float64 {
		return 0
	}

	defer func() {
		err := recover()
		if ok := assert.NotNil(t, err); !ok {
			t.Fatal()
		}
	}()

	MustPartialCall(x, "a", 10, 20)
}
