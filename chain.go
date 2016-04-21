package wfe

func init() {
	Register(chain)
}

func chain(ctx *Context, request Request, chain ...PartialRequest) interface{} {
	res, err := ctx.Apply(request)
	if err != nil {
		panic(err)
	}
	var v interface{}

	for _, ch := range chain {
		v = res.MustGet()

		ch.Append(v)
		chReq, err := ch.Request()
		if err != nil {
			panic(err)
		}

		res = ctx.MustApply(chReq)
	}

	return res.MustGet()
}

func (c *clientImpl) Chain(request Request, callbacks ...PartialRequest) (Result, error) {
	args := make([]interface{}, 0, 1+len(callbacks))
	args = append(args, request)
	for _, r := range callbacks {
		args = append(args, r)
	}

	return c.Apply(
		MustCall(chain, args...),
	)
}
