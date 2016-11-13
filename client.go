package wfe

//Client interface
type Client interface {
	//Apply a task and return a result object
	Apply(req Request) (Result, error)

	//Group creates a group of tasks. A grouped tasks are executed in parallel and returns a GroupResult object.
	//GroupResult objects can be used to wait for each tasks separately
	Group(requests ...Request) (GroupResult, error)

	//Chain creates a task chain, where the each task result is fed as argument to the next task in the chain.
	Chain(request Request, chain ...PartialRequest) (Result, error)

	/*Chord create a task chord, where a group of tasks are executed in parallel, and when all parallel tasks are
	complete successfully, all their results is fed to the callback tasks.

	The callback PartialRequest must accept same number of argument as the number of the parallel tasks, or be a variadic function
	*/
	Chord(callback PartialRequest, requests ...Request) (Result, error)

	/*
		Get a result instance for a running task knowing the task id
	*/
	ResultFor(id string) Result

	/*
		Close the client.
	*/
	Close() error
}

type clientImpl struct {
	dispatcher Dispatcher
	store      ResultStore
	parentID   string
}

//NewClient creates a new client instance.
func NewClient(o *Options) (Client, error) {
	broker, err := o.GetBroker()
	if err != nil {
		return nil, err
	}

	store, err := o.GetStore()
	if err != nil {
		return nil, err
	}

	return newClient(broker, store)
}

func newClient(broker Broker, store ResultStore) (*clientImpl, error) {
	dispatcher, err := broker.Dispatcher()

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

func (c *clientImpl) ResultFor(id string) Result {
	return &resultImpl{
		id:    id,
		store: c.store,
	}
}

func (c *clientImpl) Apply(req Request) (Result, error) {
	if c.parentID != "" && req.ParentID() == "" {
		if req, ok := req.(ParentIDSetter); ok {
			req.SetParentID(c.parentID)
		}
	}

	fn, ok := registered(req.Fn())
	if !ok {
		return nil, ErrUnknownFunction
	}

	msg := Message{
		Content: req,
	}

	o := WorkQueueRoute
	if fn.queue != "" {
		o = &RouteOptions{
			Queue:   fn.queue,
			Durable: true,
		}
	}

	id, err := c.dispatcher.Dispatch(o, &msg)
	if err != nil {
		return nil, err
	}

	result := &resultImpl{
		id:    id,
		store: c.store,
	}

	return result, nil
}
