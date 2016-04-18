package wfe

import (
	"encoding/gob"
	"sync"
)

func init() {
	Register(group)
	gob.Register([]ResultTuple{})
}

func group(c *Context, requests ...Request) ([]ResultTuple, error) {
	results := make([]ResultTuple, len(requests))
	var wg sync.WaitGroup
	wg.Add(len(results))
	for i, request := range requests {
		result, err := c.Apply(request)
		if err != nil {
			return nil, err
		}

		go func(x int, res Result) {
			defer wg.Done()

			r, e := res.Get()
			results[x] = ResultTuple{r, e}
		}(i, result)
	}

	wg.Wait()
	return results, nil
}

func Group(requests ...Request) Request {
	args := make([]interface{}, 0, len(requests))
	for _, r := range requests {
		args = append(args, r)
	}
	return MustCall(group, args...)
}
