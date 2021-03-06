package wfe

import (
	"encoding/gob"
	"errors"
	"sync"
)

func init() {
	gob.Register(interface{}(0))
}

//Result interface
type Result interface {
	//ID of the result (matches the Request ID)
	ID() string

	//Get waits for the task to finish and return the returned object and an error if the task failed
	Get() (interface{}, error)

	//MustGet same as Get but panics on error.
	MustGet() interface{}
}

type resultImpl struct {
	id    string
	store ResultStore
	o     sync.Once
	value interface{}
	err   error
}

func (r *resultImpl) ID() string {
	return r.id
}

func (r *resultImpl) Get() (interface{}, error) {
	r.o.Do(func() {
		results, err := r.get()
		r.value = results
		r.err = err
	})

	return r.value, r.err
}

func (r *resultImpl) MustGet() interface{} {
	v, e := r.Get()
	if e != nil {
		panic(e)
	}

	return v
}

func (r *resultImpl) get() (interface{}, error) {
	response, err := r.store.Get(r.ID(), DefaultTimeout)
	if err != nil {
		return nil, err
	}

	if response.State == StateError {
		return nil, errors.New(response.Error)
	}

	return response.Result, nil
}
