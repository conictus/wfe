package wfe

//Middleware interface
type Middleware interface {
	//Enter runs before the task function is executed.
	Enter(ctx *Context)

	//Exit runs after the task function is executed, in the reverse order of the Enter
	Exit(ctx *Context)
}

//middlewareStack list of middleware
type middlewareStack []Middleware

func (m middlewareStack) Enter(ctx *Context) {
	for _, mw := range m {
		mw.Enter(ctx)
	}
}

func (m middlewareStack) Exit(ctx *Context) {
	for i := len(m) - 1; i >= 0; i-- {
		m[i].Exit(ctx)
	}
}

//NOOPMiddleware implements a middleware that does nothing. It can be used as a base for a middleware in case
//you only need to implement one method (Enter, or Exit)
type NOOPMiddleware struct{}

//Enter the middleware
func (n *NOOPMiddleware) Enter(c *Context) {}

//Exit the middleware
func (n *NOOPMiddleware) Exit(c *Context) {}
