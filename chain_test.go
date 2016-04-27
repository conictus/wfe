package wfe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type chainDispatcher struct {
	testDispatcher
	err bool
	msg *Message
}

func (g *chainDispatcher) Dispatch(msg *Message) error {
	if g.err {
		return fmt.Errorf("error dispatching message")
	}

	g.msg = msg
	return nil
}

func TestClientChainSuccess(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &chainDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	r1 := MustCall(x, 1, 2)
	r2 := MustPartialCall(x, 3)
	r3 := MustPartialCall(x, 5)

	g, err := client.Chain(
		r1, r2, r3,
	)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.NotNil(t, g); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, dispatcher.msg.ID, g.ID()); !ok {
		t.Fatal()
	}

	req := dispatcher.msg.Content.(Request)

	if ok := assert.Equal(t, "github.com/conictus/wfe.chain", req.Fn()); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, []interface{}{r1, r2, r3}, req.Args()); !ok {
		t.Fatal()
	}
}

func TestClientChainError(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &chainDispatcher{
		err: true,
	}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	r1 := MustCall(x, 1, 2)
	r2 := MustPartialCall(x, 3)
	r3 := MustPartialCall(x, 5)

	_, err = client.Chain(
		r1, r2, r3,
	)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}
}

//
//func add(c *Context, a, b int) int {
//	return a + b
//}
//
//
//func TestChainTask(t *testing.T) {
//	Register(add)
//
//	broker := &testBroker{}
//	store := &testStore{}
//
//	dispatcher := &testDispatcher{}
//	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)
//
//	client, err := newClient(broker, store)
//	if ok := assert.Nil(t, err); !ok {
//		t.Fatal()
//	}
//
//	r1 := MustCall(add, 1, 2)
//	dispatcher.On("Dispatch", &Message{
//		ID:      r1.ID(),
//		Content: r1,
//	}).Return(nil)
//
//	store.On("Get", r1.ID(), DefaultTimeout).Return(&Response{
//		UUID:   r1.ID(),
//		State:  StateSuccess,
//		Result: 3,
//	}, nil)
//
//	r2 := MustPartialCall(add, 3)
//	{
//		r := MustPartialCall(add, 3)
//		r.Append(3)
//		req, _ := r.Request()
//		dispatcher.On("Dispatch", &Message{
//			ID:      req.ID(),
//			Content: req,
//		})
//	}
//
//	store.On("Get", r2.ID(), DefaultTimeout).Return(&Response{
//		UUID:   r2.ID(),
//		State:  StateSuccess,
//		Result: 6,
//	}, nil)
//
//	r3 := MustPartialCall(add, 4)
//	{
//		r := MustPartialCall(add, 4)
//		r.Append(6)
//		req, _ := r.Request()
//		dispatcher.On("Dispatch", &Message{
//			ID:      r3.ID(),
//			Content: req,
//		})
//	}
//	store.On("Get", r3.ID(), DefaultTimeout).Return(&Response{
//		UUID:   r3.ID(),
//		State:  StateSuccess,
//		Result: 10,
//	}, nil)
//
//	ctx := Context{
//		client: client,
//	}
//
//	v := chain(&ctx, r1, r2, r3)
//	if ok := assert.Equal(t, 10, v); !ok {
//		t.Fatal()
//	}
//
//	if ok := store.AssertExpectations(t); !ok {
//		t.Fatal()
//	}
//
//	if ok := dispatcher.AssertExpectations(t); !ok {
//		t.Fatal()
//	}
//}

//
//func TestGroupTaskPanic(t *testing.T) {
//	broker := &testBroker{}
//	store := &testStore{}
//
//	dispatcher := &testDispatcher{}
//	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)
//
//	client, err := newClient(broker, store)
//	if ok := assert.Nil(t, err); !ok {
//		t.Fatal()
//	}
//
//	x := func(c *Context, a, b int) int {
//		return 0
//	}
//
//	r1 := MustCall(x, 1, 2)
//	r2 := MustCall(x, 3, 4)
//	r3 := MustCall(x, 5, 6)
//
//	for _, req := range []Request{r1, r2, r3} {
//		dispatcher.On("Dispatch", &Message{
//			ID:      req.ID(),
//			Content: req,
//		}).Return(fmt.Errorf("panic for me"))
//	}
//
//	ctx := Context{
//		client: client,
//	}
//
//	defer func() {
//		err := recover()
//		if ok := assert.NotNil(t, err); !ok {
//			t.Fatal()
//		}
//	}()
//
//	group(&ctx, r1, r2, r3)
//}
