package wfe

import (
	"encoding/gob"
	"errors"
	"sync"
)

func init() {
	gob.Register(ResultTuple{})
}

type ResultTuple []interface{}

type Result interface {
	ID() string
	Get() (ResultTuple, error)
}

type resultImpl struct {
	Request
	store  ResultStore
	o      sync.Once
	values ResultTuple
	err    error
}

func (r *resultImpl) Get() (ResultTuple, error) {
	r.o.Do(func() {
		results, err := r.get()
		r.values = results
		r.err = err
	})

	return r.values, r.err
}

func (r *resultImpl) get() (ResultTuple, error) {
	response, err := r.store.Get(r.ID(), DefaultTimeout)
	if err != nil {
		return nil, err
	}

	if response.State == StateError {
		return nil, errors.New(response.Error)
	}

	return ResultTuple(response.Results), nil
}
