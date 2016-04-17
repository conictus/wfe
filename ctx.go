package wfe

type Context struct {
	Client
	id string
}

func (c *Context) UUID() string {
	return c.id
}

func (c *Context) Apply(req Request) (Result, error) {
	call := &requestImpl{
		UUID:       req.ID(),
		ParentUUID: c.id,
		Function:   req.Fn(),
		Arguments:  req.Args(),
	}

	return c.Client.Apply(call)
}
