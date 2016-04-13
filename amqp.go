package wfe

import (
	"bytes"
	"encoding/gob"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"sync"
)

const (
	amqpWrokQueue           = "wfe.work"
	amqpRequestContentType  = "application/wfe+request"
	amqpResponseContentType = "application/wfe+response"
	amqpContentEncoding     = "encoding/gob"
)

type amqpRequest struct {
	amqp.Delivery
	broker *amqpBroker
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

func (r *amqpRequest) Respond(response Response) error {
	return r.broker.Respond(r.ReplyTo, response)
}

//
//type amqpResponse struct {
//	amqp.Delivery
//}
//
//func (r *amqpResponse) ID() string {
//	return r.CorrelationId
//}
//
//func (r *amqpResponse) Values() []interface{} {
//	return nil
//}

type amqpBroker struct {
	con     *amqp.Connection
	ctx     context.Context
	invalid bool

	channel  *amqp.Channel
	callInit sync.Once

	responseCh   *amqp.Channel
	responseInit sync.Once
}

func NewAMQPBroker(url string) (Broker, error) {
	var broker amqpBroker
	if err := broker.init(url); err != nil {
		return nil, err
	}
	return &broker, nil
}

func (b *amqpBroker) init(url string) error {
	c, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	n := make(chan *amqp.Error)
	c.NotifyClose(n)
	go func() {
		//defer close(n)
		err := <-n
		log.Errorf("Connection closed: %s", err)
		b.invalid = true
	}()

	b.con = c
	return nil
}

func (b *amqpBroker) initChannel() error {
	var err error
	b.callInit.Do(func() {
		b.channel, err = b.con.Channel()
		if err != nil {
			return
		}
	})

	return err
}

func (b *amqpBroker) Call(call Call) error {
	if err := b.initChannel(); err != nil {
		return err
	}

	if _, err := b.channel.QueueDeclare(amqpWrokQueue, true, false, false, false, nil); err != nil {
		return err
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(call); err != nil {
		return err
	}

	return b.channel.Publish("", amqpWrokQueue, false, false, amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     amqpRequestContentType,
		ContentEncoding: amqpContentEncoding,
		ReplyTo:         "test.results",
		Body:            buffer.Bytes(),
		CorrelationId:   call.UUID,
	})
}

func (b *amqpBroker) Respond(queue string, response Response) error {
	if err := b.initChannel(); err != nil {
		return err
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(response); err != nil {
		return err
	}

	return b.channel.Publish("", queue, false, false, amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     amqpResponseContentType,
		ContentEncoding: amqpContentEncoding,
		Body:            buffer.Bytes(),
		CorrelationId:   response.UUID,
	})
}

func (b *amqpBroker) Requests() (<-chan Request, error) {
	ch, err := b.con.Channel()
	if err != nil {
		return nil, err
	}

	if _, err = ch.QueueDeclare(amqpWrokQueue, true, false, false, false, nil); err != nil {
		return nil, err
	}

	msges, err := ch.Consume(amqpWrokQueue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	feeder := make(chan Request)

	go func() {
		defer close(feeder)
		for msg := range msges {
			if msg.ContentType != amqpRequestContentType {
				log.Warningf("received a request with wrong content type '%s', ignoring.", msg.ContentType)
				break
			}
			feeder <- &amqpRequest{
				Delivery: msg,
				broker:   b,
			}
		}
	}()

	return feeder, nil
}

func (b *amqpBroker) Responses(queue string) (<-chan Response, error) {
	ch, err := b.con.Channel()
	if err != nil {
		return nil, err
	}

	if _, err = ch.QueueDeclare(queue, false, true, true, false, nil); err != nil {
		return nil, err
	}

	msges, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	feeder := make(chan Response)

	go func() {
		defer close(feeder)
		for msg := range msges {
			if msg.ContentType != amqpResponseContentType {
				log.Warningf("received a response with wrong content type '%s', ignoring.", msg.ContentType)
				break
			}
			var response Response
			buffer := bytes.NewBuffer(msg.Body)

			decoder := gob.NewDecoder(buffer)
			if err := decoder.Decode(&response); err != nil {
				log.Errorf("Failed to decoder response message: %s", err)
				continue
			}
			log.Debugf("Got response %s", response)
			feeder <- response
		}
	}()

	return feeder, nil
}

func (b *amqpBroker) Close() {
	b.con.Close()
}
