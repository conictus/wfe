package wfe

type Client interface {
	Apply(req Request) (Result, error)
	Close() error
}

type clientImpl struct {
	dispatcher Dispatcher
	store      ResultStore
}

func NewClient(o *Options) (Client, error) {
	broker, err := o.GetBroker()
	if err != nil {
		return nil, err
	}

	store, err := o.GetStore()
	if err != nil {
		return nil, err
	}

	dispatcher, err := broker.Dispatcher(WorkQueueRoute)

	if err != nil {
		return nil, err
	}

	return &clientImpl{
		dispatcher: dispatcher,
		store:      store,
	}, nil
}

func (c *clientImpl) Close() error {
	return c.dispatcher.Close()
}

func (c *clientImpl) Apply(req Request) (Result, error) {
	msg := Message{
		ID:      req.ID(),
		Content: req,
	}

	if err := c.dispatcher.Dispatch(&msg); err != nil {
		return nil, err
	}

	result := &resultImpl{
		Request: req,
		store:   c.store,
	}

	return result, nil
}
