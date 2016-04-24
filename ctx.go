package wfe

//Context is always the first argument to a Task function. It's mainly used by a task to start and spawn other tasks.
//Context implements the Client interface.
type Context struct {
	Client
	id string
}

//UUID of the current task.
func (c *Context) UUID() string {
	return c.id
}

//MustApply same as Apply but panics on error.
func (c *Context) MustApply(req Request) Result {
	r, e := c.Apply(req)
	if e != nil {
		panic(e)
	}
	return r
}
