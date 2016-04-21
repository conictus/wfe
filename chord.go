package wfe

func init() {
	Register(chord)
}

func chord(c *Context, callback PartialRequest, requests ...Request) interface{} {
	g, err := c.Group(requests...)
	if err != nil {
		panic(err)
	}

	for i := 0; i < g.Count(); i++ {
		r, err := g.ResultOf(i)
		if err != nil {
			panic(err)
		}
		v, err := r.Get()
		if err != nil {
			panic(err)
		}

		callback.Append(v)
	}

	req, err := callback.Request()
	if err != nil {
		panic(err)
	}

	return c.MustApply(req).MustGet()
}

func (c *clientImpl) Chord(callback PartialRequest, requests ...Request) (Result, error) {
	args := make([]interface{}, 0, 1+len(requests))
	args = append(args, callback)
	for _, r := range requests {
		args = append(args, r)
	}

	return c.Apply(
		MustCall(chord, args...),
	)
}
