package wfe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type groupDispatcher struct {
	testDispatcher

	msg *Message
}

func (g *groupDispatcher) Dispatch(msg *Message) error {
	g.msg = msg
	return nil
}

func TestClientGroupSuccess(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &groupDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

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

	//req := MustCall(group, r1, r2, r3)
	//msg := Message{
	//	ID:      req.ID(),
	//	Content: req,
	//}

	//dispatcher.On("Dispatch", msg).Return(nil)

	g, err := client.Group(
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

	if ok := assert.Equal(t, "github.com/conictus/wfe.group", req.Fn()); !ok {
		t.Fatal()
	}
}
