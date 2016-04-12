package wfe

import "github.com/op/go-logging"

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
	call, err := request.Call()
	if err != nil {
		log.Errorf("Failed to get request call: %s", err)
		return nil //we return nil to make sure we discard this corrupted message
	}

	log.Debugf("Received a message: %s", call)
	return nil
}

func (e *Engine) Run() error {
	requests, err := e.broker.Consume()
	if err != nil {
		return err
	}

	for request := range requests {
		if err := e.dispatch(request); err != nil {
			request.Ack()
		}
	}

	return nil
}
