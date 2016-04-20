package wfe

type Context struct {
	Client
	id string
}

func (c *Context) UUID() string {
	return c.id
}

func (c *Context) setParent(req Request) Request {
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

func (c *Context) Group(requests ...Request) (GroupResult, error) {
	reqs := make([]Request, 0, len(requests))
	for _, req := range requests {
		reqs = append(reqs, c.setParent(req))
	}

	return c.Client.Group(reqs...)
}
