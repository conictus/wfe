package wfe

import (
	"encoding/gob"
	"fmt"
)

func init() {
	gob.Register([]interface{}{})
	Register(group)
}

type GroupResult interface {
	Result
	ResultOf(i int) (Result, error)
	Count() int
}

type groupResultImpl struct {
	Result
	store ResultStore
}

func (g *groupResultImpl) get() ([]string, error) {
	r, err := g.Get()
	if err != nil {
		return nil, err
	}

	if ids, ok := r.([]string); ok {
		return ids, nil
	}

	return nil, fmt.Errorf("invalid group result")
}

func (g *groupResultImpl) Count() int {
	ids, _ := g.get()
	return len(ids)
}

func (g *groupResultImpl) ResultOf(i int) (Result, error) {
	ids, err := g.get()
	if err != nil {
		return nil, err
	}

	if i >= len(ids) {
		return nil, fmt.Errorf("index out of range")
	}

	return &resultImpl{
		id:    ids[i],
		store: g.store,
	}, nil
}

func group(c *Context, requests ...Request) []string {
	results := make([]string, len(requests))
	for i, request := range requests {
		result, err := c.Apply(request)
		if err != nil {
			panic(err)
		}

		results[i] = result.ID()
	}

	return results
}

func (c *clientImpl) Group(requests ...Request) (GroupResult, error) {
	args := make([]interface{}, 0, len(requests))
	for _, r := range requests {
		args = append(args, r)
	}
	request := MustCall(group, args...)
	result, err := c.Apply(request)
	if err != nil {
		return nil, err
	}

	return &groupResultImpl{
		Result: result,
		store:  c.store,
	}, nil
}
