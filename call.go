package wfe

type Call struct {
	UUID      string
	Function  string
	Arguments []interface{}
}

func (c *Call) ID() string {
	return c.UUID
}

func (c *Call) Get() []interface{} {
	return nil
}
