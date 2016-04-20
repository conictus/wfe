package wfe

import (
	"encoding/gob"
)

func init() {
	gob.Register([]interface{}{})
	Register(group)
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

func Group(requests ...Request) Request {
	args := make([]interface{}, 0, len(requests))
	for _, r := range requests {
		args = append(args, r)
	}
	return MustCall(group, args...)
}
