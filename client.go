package wfe

import (
	"errors"
	"fmt"
	"github.com/pborman/uuid"
)

var (
	TooFewArgumentsError  = errors.New("call with too few arguments")
	TooManyArgumentsError = errors.New("call with too many arguments")
)

type Client interface {
	Apply(req Request) (Result, error)
}

type clientImpl struct {
	dispatcher Dispatcher
	consumer   Consumer
	replyTo    string

	results map[string]chan *Response
}

func NewClient(broker Broker) (Client, error) {
	dispatcher, err := broker.Dispatcher(WorkQueueRoute)

	if err != nil {
		return nil, err
	}
	replyTo := fmt.Sprintf("client.%s", uuid.New())
	consumer, err := broker.Consumer(&RouteOptions{
		Queue:       replyTo,
		Exclusive:   true,
		AutoConfirm: true,
		AutoDelete:  true,
	})

	if err != nil {
		return nil, err
	}

	c := &clientImpl{
		replyTo:    replyTo,
		dispatcher: dispatcher,
		consumer:   consumer,
		results:    make(map[string]chan *Response),
	}

	c.receiveResponses()
	return c, nil
}

func (c *clientImpl) Close() {
	c.consumer.Close()
	c.dispatcher.Close()
}

func (c *clientImpl) dispatchResponse(id string, r *Response) {
	defer delete(c.results, id)
	if ch, ok := c.results[id]; ok {
		defer close(ch)
		ch <- r
	} else {
		log.Warningf("Received unkonwn response id: %s", id)
	}
}

func (c *clientImpl) receiveResponses() {
	go func() {
		deliveries, err := c.consumer.Consume()
		if err != nil {
			log.Errorf("Failed to receive responses:", err)
			return
		}

		for delivery := range deliveries {
			log.Debugf("Received response: %s", delivery.ID())
			var response Response
			if err := delivery.Content(&response); err != nil {
				response.State = StateError
				response.Error = err.Error()
			}

			c.dispatchResponse(delivery.ID(), &response)
		}
	}()
}

func (c *clientImpl) Apply(req Request) (Result, error) {
	msg := Message{
		ID:      req.ID(),
		ReplyTo: c.replyTo,
		Content: req,
	}

	if err := c.dispatcher.Dispatch(&msg); err != nil {
		return nil, err
	}

	result := &resultImpl{
		Request: req,
		ch:      make(chan *Response, 1),
	}

	c.results[result.ID()] = result.ch

	return result, nil
}
