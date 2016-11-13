package wfe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type groupDispatcher struct {
	testDispatcher
	err bool
	msg *Message
}

func (g *groupDispatcher) Dispatch(o *RouteOptions, msg *Message) (string, error) {
	if g.err {
		return "", fmt.Errorf("error dispatching message")
	}

	g.msg = msg
	return "", nil
}

func TestClientGroupSuccess(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &groupDispatcher{}
	broker.On("Dispatcher").Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	g, err := client.Group(
		r1, r2, r3,
	)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.NotNil(t, g); !ok {
		t.Fatal()
	}

	req := dispatcher.msg.Content.(Request)

	if ok := assert.Equal(t, "github.com/conictus/wfe.group", req.Fn()); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, []interface{}{r1, r2, r3}, req.Args()); !ok {
		t.Fatal()
	}
}

func TestClientGroupError(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &groupDispatcher{
		err: true,
	}
	broker.On("Dispatcher").Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	_, err = client.Group(
		r1, r2, r3,
	)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}

}

func TestGroupTask(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher").Return(dispatcher, nil)

	client, err := newClient(broker, store)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	Register(x)

	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	for i, req := range []Request{r1, r2, r3} {
		dispatcher.On("Dispatch", WorkQueueRoute, &Message{
			Content: req,
		}).Return(fmt.Sprintf("id-%d", i), nil)
	}

	ctx := Context{
		client: client,
	}

	ids := group(&ctx, r1, r2, r3)

	if ok := dispatcher.AssertNumberOfCalls(t, "Dispatch", 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, []string{"id-0", "id-1", "id-2"}, ids); !ok {
		t.Fatal()
	}

	for _, req := range []Request{r1, r2, r3} {
		if ok := dispatcher.AssertCalled(t, "Dispatch", WorkQueueRoute, &Message{
			Content: req,
		}); !ok {
			t.Fatal()
		}
	}
}

func TestGroupTaskPanic(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher").Return(dispatcher, nil)

	client, err := newClient(broker, store)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	for _, req := range []Request{r1, r2, r3} {
		dispatcher.On("Dispatch", WorkQueueRoute, &Message{
			Content: req,
		}).Return(fmt.Errorf("panic for me"))
	}

	ctx := Context{
		client: client,
	}

	defer func() {
		err := recover()
		if ok := assert.NotNil(t, err); !ok {
			t.Fatal()
		}
	}()

	group(&ctx, r1, r2, r3)
}
