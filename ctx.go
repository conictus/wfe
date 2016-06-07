package wfe

import "fmt"

//Context is always the first argument to a Task function. It's mainly used by a task to start and spawn other tasks.
//Context implements the Client interface.
type Context struct {
	client Client
	id     string

	values map[string]interface{}
}

//UUID of the current task.
func (c *Context) UUID() string {
	return c.id
}

//Set store a variable on the ctx. Should be used by middleware to inject values to tasks
func (c *Context) Set(name string, value interface{}) {
	c.values[name] = value
}

//Get returns a value previously stored by Set.
func (c *Context) Get(name string) (interface{}, bool) {
	v, ok := c.values[name]
	return v, ok
}

//MustGet like Get but panics if name is not found.
func (c *Context) MustGet(name string) interface{} {
	if v, ok := c.Get(name); !ok {
		panic(fmt.Errorf("key not found '%s'", name))
	} else {
		return v
	}
}

//MustApply same as Apply but panics on error.
func (c *Context) MustApply(req Request) Result {
	r, e := c.Apply(req)
	if e != nil {
		panic(e)
	}
	return r
}

//Apply a task and return a result object
func (c *Context) Apply(req Request) (Result, error) {
	return c.client.Apply(req)
}

//Group creates a group of tasks. A grouped tasks are executed in parallel and returns a GroupResult object.
//GroupResult objects can be used to wait for each tasks separately
func (c *Context) Group(requests ...Request) (GroupResult, error) {
	return c.client.Group(requests...)
}

//Chain creates a task chain, where the each task result is fed as argument to the next task in the chain.
func (c *Context) Chain(request Request, chain ...PartialRequest) (Result, error) {
	return c.client.Chain(request, chain...)
}

/*Chord create a task chord, where a group of tasks are executed in parallel, and when all parallel tasks are
complete successfully, all their results is fed to the callback tasks.

The callback PartialRequest must accept same number of argument as the number of the parallel tasks, or be a variadic function
*/
func (c *Context) Chord(callback PartialRequest, requests ...Request) (Result, error) {
	return c.client.Chord(callback, requests...)
}

/*
ResultFor Gets a result instance for a running task knowing the task id
*/
func (c *Context) ResultFor(id string) Result {
	return c.client.ResultFor(id)
}

func (c *Context) Close() error {
	//do nothing
	return nil
}
