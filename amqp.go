package wfe

import (
	"bytes"
	"encoding/gob"
	"github.com/streadway/amqp"
	"sync"
)

type amqpRequest struct {
	amqp.Delivery
}

func (r *amqpRequest) Ack() error {
	return r.Delivery.Ack(false)
}

func (r *amqpRequest) Call() (Call, error) {
	//un serialize the body and return a valid call
	decoder := gob.NewDecoder(bytes.NewBuffer(r.Body))
	var c Call
	if err := decoder.Decode(&c); err != nil {
		return c, err
	}

	return c, nil
}

type amqpBroker struct {
	c    *amqp.Connection
	disp *amqp.Channel
	cons *amqp.Channel

	dispInit sync.Once
	consInit sync.Once
}

func NewAMQPBroker(url string) (Broker, error) {
	con, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	broker := &amqpBroker{
		c: con,
	}

	return broker, nil
}

func (b *amqpBroker) initDispatch() error {
	var err error
	b.dispInit.Do(func() {
		b.disp, err = b.c.Channel()
		if err != nil {
			return
		}

		if _, err = b.disp.QueueDeclare(WorkQueue, true, false, false, false, nil); err != nil {
			return
		}
	})

	return err
}

func (b *amqpBroker) initConsume() error {
	var err error
	b.dispInit.Do(func() {
		b.cons, err = b.c.Channel()
		if err != nil {
			return
		}

		if _, err = b.cons.QueueDeclare(WorkQueue, true, false, false, false, nil); err != nil {
			return
		}
	})

	return err
}

func (b *amqpBroker) Dispatch(call Call) error {
	if err := b.initDispatch(); err != nil {
		return err
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(call); err != nil {
		return err
	}

	return b.disp.Publish("", WorkQueue, false, false, amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     contentType,
		ContentEncoding: contentEncoding,
		Body:            buffer.Bytes(),
		CorrelationId:   call.UUID,
	})
}

func (b *amqpBroker) Consume() (<-chan Request, error) {
	if err := b.initConsume(); err != nil {
		return nil, err
	}

	msges, err := b.cons.Consume(WorkQueue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	ch := make(chan Request)

	go func() {
		for msg := range msges {
			if msg.ContentType != contentType {
				log.Warning("received a message with wrong content type '%s', ignoring.", msg.ContentType)
				continue
			}
			ch <- &amqpRequest{
				msg,
			}
		}
	}()

	return ch, nil
}
