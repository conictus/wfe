package wfe

type Context struct {
	id     string
	parent string
}

func (c *Context) UUID() string {
	return c.id
}

func (c *Context) ParentUUID() string {
	return c.parent
}

func (c *Context) Call(work interface{}, args ...interface{}) (Response, error) {
	return nil, nil
}
