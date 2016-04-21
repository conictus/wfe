package wfe

type Context struct {
	Client
	id string
}

func (c *Context) UUID() string {
	return c.id
}

func (c *Context) setParent(req Request) *requestImpl {
	return &requestImpl{
		UUID:       req.ID(),
		ParentUUID: c.id,
		Function:   req.Fn(),
		Arguments:  req.Args(),
	}
}

func (c *Context) Apply(req Request) (Result, error) {
	return c.Client.Apply(c.setParent(req))
}

func (c *Context) MustApply(req Request) Result {
	r, e := c.Apply(req)
	if e != nil {
		panic(e)
	}
	return r
}

func (c *Context) Group(requests ...Request) (GroupResult, error) {
	reqs := make([]Request, 0, len(requests))
	for _, req := range requests {
		reqs = append(reqs, c.setParent(req))
	}

	return c.Client.Group(reqs...)
}

func (c *Context) Chain(request Request, chain ...PartialRequest) (Result, error) {
	chainReq := make([]PartialRequest, 0, len(chain))
	for _, req := range chain {
		chainReq = append(chainReq, c.setParent(req))
	}

	return c.Client.Chain(c.setParent(request), chainReq...)
}

func (c *Context) Chord(callback PartialRequest, requests ...Request) (Result, error) {
	reqs := make([]Request, 0, len(requests))
	for _, req := range requests {
		reqs = append(reqs, c.setParent(req))
	}

	return c.Client.Chord(c.setParent(callback), reqs...)
}
