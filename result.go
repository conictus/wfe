package wfe

import (
	"errors"
	"sync"
)

type Result interface {
	ID() string
	Get() ([]interface{}, error)
}

type resultImpl struct {
	Request
	store  ResultStore
	o      sync.Once
	values []interface{}
	err    error
}

func (r *resultImpl) Get() ([]interface{}, error) {
	r.o.Do(func() {
		results, err := r.get()
		r.values = results
		r.err = err
	})

	return r.values, r.err
}

func (r *resultImpl) get() ([]interface{}, error) {
	response, err := r.store.Get(r.ID(), DefaultTimeout)
	if err != nil {
		return nil, err
	}

	if response.State == StateError {
		return nil, errors.New(response.Error)
	}

	return response.Results, nil
}
