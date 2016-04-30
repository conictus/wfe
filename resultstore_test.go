package wfe

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisStoreInit(t *testing.T) {
	o := Options{
		Store: "redis://localhost:6379",
	}

	store, err := o.GetStore()

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Implements(t, (*ResultStore)(nil), store); !ok {
		t.Fatal()
	}
}

func TestRedisStoreSetGet(t *testing.T) {
	o := Options{
		Store: "redis://localhost:6379",
	}

	store, err := o.GetStore()

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	r := &Response{
		UUID:   "1234",
		State:  StateSuccess,
		Error:  "error",
		Result: 123,
	}

	err = store.Set(r)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	resp, err := store.Get("1234", DefaultTimeout)

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, r, resp); !ok {
		t.Fatal()
	}
}

func TestRedisStoreGetTimeout(t *testing.T) {
	o := Options{
		Store: "redis://localhost:6379",
	}

	store, err := o.GetStore()

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	_, err = store.Get("does not exist", 2)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, ErrTimeout, err); !ok {
		t.Fatal()
	}
}

func TestRedisStoreDefaultTimeout(t *testing.T) {
	o := Options{
		Store: "redis://localhost:6379?timeout=3",
	}

	store, err := o.GetStore()

	if ok := assert.Nil(t, err); !ok {
		t.Fatal()
	}

	ts := time.Now()
	_, err = store.Get("does not exist", DefaultTimeout)

	if ok := assert.Error(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, ErrTimeout, err); !ok {
		t.Fatal()
	}
	d := time.Since(ts)
	//we increased the delta, because codeship redis apparently is very slow
	delta := 2 * time.Second

	if ok := assert.InDelta(t, int64(3*time.Second), int64(d), float64(delta)); !ok {
		t.Fatal()
	}
}
