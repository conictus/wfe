package wfe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func wfeAddTest(c *Context, a, b int) int {
	return a + b
}

type wfeTestStruct struct {
	s string
}

func (w *wfeTestStruct) String() string {
	return w.s
}

func wfeTestPtr(c *Context, a *wfeTestStruct) string {
	return a.s
}

func wfeTestInter(c *Context, a fmt.Stringer) string {
	return a.String()
}

func TestHandleRequestUnknownFn(t *testing.T) {
	eng := &Engine{}

	req := MustCall(wfeAddTest, 1, 2)
	_, err := eng.handle(req)

	if ok := assert.Equal(t, err, ErrUnknownFunction); !ok {
		t.Fatal()
	}
}

func TestHandleRequestOk(t *testing.T) {
	Register(wfeAddTest)
	eng := &Engine{}

	req := MustCall(wfeAddTest, 1, 2)
	v, err := eng.handle(req)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 3, v); !ok {
		t.Fatal()
	}
}

func TestHandleRequestPtrOk(t *testing.T) {
	t.Skip("handle request with pointer is broken")
	Register(wfeTestPtr)
	eng := &Engine{}

	x := wfeTestStruct{"test string"}
	req := requestImpl{
		Function:  "github.com/conictus/wfe.wfeTestPtr",
		Arguments: []interface{}{x},
	}
	v, err := eng.handle(&req)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, x.s, v); !ok {
		t.Fatal()
	}
}

func TestHandleRequestInterOk(t *testing.T) {
	Register(wfeTestInter)
	eng := &Engine{}

	x := wfeTestStruct{"test string"}
	req := requestImpl{
		Function:  "github.com/conictus/wfe.wfeTestInter",
		Arguments: []interface{}{x},
	}
	v, err := eng.handle(&req)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, x.s, v); !ok {
		t.Fatal()
	}
}
