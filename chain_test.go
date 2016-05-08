package wfe

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

//
//type chainDispatcher struct {
//	testDispatcher
//}
//
//func (g *chainDispatcher) Dispatch(msg *Message) (error) {
//	args := g.Called(msg)
//	return args.Error(0)
//}

func TestClientChainSuccess(t *testing.T) {
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

	r1 := MustCall(x, 1, 2)
	r2 := MustPartialCall(x, 3)
	r3 := MustPartialCall(x, 5)

	dispatcher.On("Dispatch", &Message{
		Content: &requestImpl{
			Function:  "github.com/conictus/wfe.chain",
			Arguments: []interface{}{r1, r2, r3},
		},
	}).Return("", nil)

	g, err := client.Chain(
		r1, r2, r3,
	)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.NotNil(t, g); !ok {
		t.Fatal()
	}

	if ok := dispatcher.AssertExpectations(t); !ok {
		t.Fatal()
	}
}

func TestClientChainError(t *testing.T) {
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

	r1 := MustCall(x, 1, 2)
	r2 := MustPartialCall(x, 3)
	r3 := MustPartialCall(x, 5)

	dispatcher.On("Dispatch", &Message{
		Content: &requestImpl{
			Function:  "github.com/conictus/wfe.chain",
			Arguments: []interface{}{r1, r2, r3},
		},
	}).Return("", errors.New("die for me"))

	g, err := client.Chain(
		r1, r2, r3,
	)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Nil(t, g); !ok {
		t.Fatal()
	}

	if ok := dispatcher.AssertExpectations(t); !ok {
		t.Fatal()
	}
}
