package wfe

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	broker.On("Dispatcher", WorkQueueRoute).Return(&testDispatcher{}, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := broker.AssertCalled(t, "Dispatcher", WorkQueueRoute); !ok {
		t.Fatal()
	}

	if ok := assert.Implements(t, (*Client)(nil), client); !ok {
		t.Fatal()
	}
}

func TestClientApplySuccess(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	req := MustCall(x, 1, 2)
	msg := Message{
		ID:      req.ID(),
		Content: req,
	}

	dispatcher.On("Dispatch", &msg).Return(nil)
	result, err := client.Apply(req)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := dispatcher.AssertExpectations(t); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, req.ID(), result.ID()); !ok {
		t.Fatal()
	}
}

func TestClientApplyError(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	req := MustCall(x, 1, 2)
	msg := Message{
		ID:      req.ID(),
		Content: req,
	}

	dispatcher.On("Dispatch", &msg).Return(errors.New("stupid error"))
	_, err = client.Apply(req)
	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}
}

func TestClientApplyParentIDInjection(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)
	client.parentID = "i am your father"
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return 0
	}

	req := MustCall(x, 1, 2)
	msg := Message{
		ID:      req.ID(),
		Content: req,
	}

	dispatcher.On("Dispatch", &msg).Return(nil)
	_, err = client.Apply(req)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, client.parentID, req.ParentID()); !ok {
		t.Fatal()
	}
}

func TestClientResultOf(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &testDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)
	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	id := "some id"
	res := client.ResultFor(id)

	if ok := assert.Equal(t, id, res.ID()); !ok {
		t.Fatal()
	}

	if res, ok := res.(*resultImpl); ok {
		if ok := assert.Equal(t, client.store, res.store); !ok {
			t.Fatal()
		}
	} else {
		t.Fatal("Not resultImpl")
	}
}
