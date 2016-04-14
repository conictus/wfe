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
	Call
	ch     chan *Response
	o      sync.Once
	values []interface{}
	err    error
}

func (r *resultImpl) ID() string {
	return r.UUID
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
	//todo: a timeout would be nice
	response := <-r.ch
	if response.State == StateError {
		return nil, errors.New(response.Error)
	}

	return response.Results, nil
}
