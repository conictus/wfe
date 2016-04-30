package wfe

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResultGetOK(t *testing.T) {
	store := &testStore{}

	store.On("Get", "1234", DefaultTimeout).Return(&Response{
		UUID:   "1234",
		State:  StateSuccess,
		Result: 10,
	}, nil)

	res := resultImpl{
		store: store,
		id:    "1234",
	}

	v, e := res.Get()
	if ok := assert.Nil(t, e); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 10, v); !ok {
		t.Fatal()
	}
}

func TestResultGetStoreError(t *testing.T) {
	store := &testStore{}

	store.On("Get", "1234", DefaultTimeout).Return(nil, errors.New("store error"))

	res := resultImpl{
		store: store,
		id:    "1234",
	}

	_, e := res.Get()
	if ok := assert.EqualError(t, e, "store error"); !ok {
		t.Fatal()
	}
}

func TestResultGetTaskError(t *testing.T) {
	store := &testStore{}

	store.On("Get", "1234", DefaultTimeout).Return(&Response{
		UUID:  "1234",
		State: StateError,
		Error: "invalid",
	}, nil)

	res := resultImpl{
		store: store,
		id:    "1234",
	}

	_, e := res.Get()
	if ok := assert.EqualError(t, e, "invalid"); !ok {
		t.Fatal()
	}

}
