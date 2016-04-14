package wfe

type Call struct {
	UUID      string
	Function  string
	Arguments []interface{}
}

type Response struct {
	UUID   string
	Values []interface{}
}

func (c *Call) Get() Result {
	return nil
}

func (c *Call) ID() string {
	return c.UUID
}
