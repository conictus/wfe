package wfe

import "github.com/op/go-logging"

type Job interface {
	UUID() string
	ParentUUID() string
	Wait()
	State() string
}

var (
	log = logging.MustGetLogger("wfe")
)

type Engine struct {
	broker Broker
}

func New(broker Broker) *Engine {
	return &Engine{
		broker: broker,
	}
}

func (e *Engine) dispatch(request CallRequest) error {
	return nil
}

func (e *Engine) Run() error {
	requests, err := e.broker.Consume()
	if err != nil {
		return err
	}

	for request := range requests {
		log.Debugf("Received a message: %s", request.Call())
		if err := e.dispatch(request); err != nil {
			request.Ack()
		}
	}

	return nil
}
