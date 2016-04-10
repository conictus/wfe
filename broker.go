package wfe

import (
	"bytes"
	"encoding/gob"
	"github.com/streadway/amqp"
)

const (
	WorkQueue       = "wfe.work"
	contentType     = "application/wfe+call"
	contentEncoding = "encoding/gob"
)

type Call struct {
	Function  string
	Arguments []interface{}
}

type callRequest struct {
	amqp.Delivery
}

type CallRequest interface {
	Ack()
	Call() Call
}

func (r *callRequest) Ack() {
	r.Delivery.Ack(false)
}

func (r *callRequest) Call() Call {
	//un serialize the body and return a valid call
	decoder := gob.NewDecoder(bytes.NewBuffer(r.Body))
	var c Call
	if err := decoder.Decode(&c); err != nil {
		log.Errorf("Failed to decode call message: %s", err)
	}

	return c
}

type Broker interface {
	Dispatch(call Call) error
	Consume() (<-chan CallRequest, error)
}

type rabbitMqBroker struct {
	con *amqp.Connection
	ch  *amqp.Channel
}

func NewRabbitMQBroker(url string) (Broker, error) {
	con, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	broker := &rabbitMqBroker{
		con: con,
	}

	if err := broker.init(); err != nil {
		return nil, err
	}

	return broker, nil
}

func (b *rabbitMqBroker) init() error {
	ch, err := b.con.Channel()
	if err != nil {
		return err
	}

	b.ch = ch
	if _, err := b.ch.QueueDeclare(WorkQueue, true, false, false, false, nil); err != nil {
		return err
	}

	return nil
}

func (b *rabbitMqBroker) Dispatch(call Call) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(call); err != nil {
		return err
	}

	return b.ch.Publish("", WorkQueue, false, false, amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     contentType,
		ContentEncoding: contentEncoding,
		Body:            buffer.Bytes(),
	})
}

func (b *rabbitMqBroker) Consume() (<-chan CallRequest, error) {
	msges, err := b.ch.Consume(WorkQueue, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	ch := make(chan CallRequest)

	go func() {
		for msg := range msges {
			if msg.ContentType != contentType {
				log.Warning("received a message with wrong content type '%s', ignoring.", msg.ContentType)
				continue
			}

			ch <- &callRequest{
				msg,
			}
		}
	}()

	return ch, nil
}
