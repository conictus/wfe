package wfe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func wfeTestPanic(c *Context) {
	panic("i paniced")
}

func TestHandleRequestUnknownFn(t *testing.T) {
	eng := &Engine{}

	req := MustCall(wfeAddTest, 1, 2)
	_, err := eng.handle("", req)

	if ok := assert.Equal(t, err, ErrUnknownFunction); !ok {
		t.Fatal()
	}
}

func TestHandleRequestOk(t *testing.T) {
	Register(wfeAddTest)
	eng := &Engine{}

	req := MustCall(wfeAddTest, 1, 2)
	v, err := eng.handle("", req)

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
	v, err := eng.handle("", &req)

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
	v, err := eng.handle("", &req)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, x.s, v); !ok {
		t.Fatal()
	}
}

type testDelivery struct {
	val requestImpl
	mock.Mock
}

func (d *testDelivery) ID() string {
	args := d.Called()
	return args.String(0)
}

func (d *testDelivery) Confirm() error {
	args := d.Called()
	return args.Error(0)
}

func (d *testDelivery) Content(c interface{}) error {
	val := c.(*requestImpl)
	*val = d.val
	//val.Function = "github.com/conictus/wfe.wfeAddTest"
	//val.Arguments = []interface{}{1, 2}
	return nil
}

func TestHandleDeliverOk(t *testing.T) {
	store := &testStore{}
	eng := &Engine{store: store}

	d := testDelivery{val: requestImpl{
		Function:  "github.com/conictus/wfe.wfeAddTest",
		Arguments: []interface{}{1, 2},
	}}

	d.On("ID").Return("1234")
	d.On("Confirm").Return(nil)

	store.On("Set", &Response{
		UUID:   "1234",
		State:  StateSuccess,
		Result: 3,
	}).Return(nil)

	err := eng.handleDelivery(&d)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := d.AssertExpectations(t); !ok {
		t.Fatal()
	}

	if ok := store.AssertExpectations(t); !ok {
		t.Fatal()
	}
}

func TestHandleDeliverHandleErr(t *testing.T) {
	store := &testStore{}
	eng := &Engine{store: store}

	d := testDelivery{val: requestImpl{
		Function:  "github.com/conictus/wfe.unknowFunctions",
		Arguments: []interface{}{1, 2},
	}}

	d.On("ID").Return("1234")
	d.On("Confirm").Return(nil)

	store.On("Set", &Response{
		UUID:  "1234",
		State: StateError,
		Error: ErrUnknownFunction.Error(),
	}).Return(nil)

	err := eng.handleDelivery(&d)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}

	if ok := d.AssertExpectations(t); !ok {
		t.Fatal()
	}

	if ok := store.AssertExpectations(t); !ok {
		t.Fatal()
	}
}

func TestHandleDeliverHandleFnPanic(t *testing.T) {
	Register(wfeTestPanic)
	store := &testStore{}
	eng := &Engine{store: store}

	d := testDelivery{val: requestImpl{
		Function:  "github.com/conictus/wfe.wfeTestPanic",
		Arguments: []interface{}{},
	}}

	d.On("ID").Return("1234")
	d.On("Confirm").Return(nil)

	store.On("Set", &Response{
		UUID:  "1234",
		State: StateError,
		Error: "i paniced",
	}).Return(nil)

	err := eng.handleDelivery(&d)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := d.AssertExpectations(t); !ok {
		t.Fatal()
	}

	if ok := store.AssertExpectations(t); !ok {
		t.Fatal()
	}
}
