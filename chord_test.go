package wfe

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type chordDispatcher struct {
	testDispatcher
}

func (g *chordDispatcher) Dispatch(msg *Message) error {
	msg.ID = ""
	req := msg.Content.(*requestImpl)
	req.UUID = ""

	args := g.Called(msg)
	return args.Error(0)
}

func TestClientChordSuccess(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &chordDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return a + b
	}

	y := func(c *Context, a, b, d int) int {
		return a + b + d
	}

	cb := MustPartialCall(y)
	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	dispatcher.On("Dispatch", &Message{
		Content: &requestImpl{
			Function:  "github.com/conictus/wfe.chord",
			Arguments: []interface{}{cb, r1, r2, r3},
		},
	}).Return(nil)

	g, err := client.Chord(
		cb,
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

func TestClientChordError(t *testing.T) {
	broker := &testBroker{}
	store := &testStore{}

	dispatcher := &chordDispatcher{}
	broker.On("Dispatcher", WorkQueueRoute).Return(dispatcher, nil)

	client, err := newClient(broker, store)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	x := func(c *Context, a, b int) int {
		return a + b
	}

	y := func(c *Context, a, b, d int) int {
		return a + b + d
	}

	cb := MustPartialCall(y)
	r1 := MustCall(x, 1, 2)
	r2 := MustCall(x, 3, 4)
	r3 := MustCall(x, 5, 6)

	dispatcher.On("Dispatch", &Message{
		Content: &requestImpl{
			Function:  "github.com/conictus/wfe.chord",
			Arguments: []interface{}{cb, r1, r2, r3},
		},
	}).Return(errors.New("die hard"))

	g, err := client.Chord(
		cb,
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
