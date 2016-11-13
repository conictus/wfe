package wfe

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
	"net"
	"net/url"
	"time"
)

const (
	amqpContentType     = "application/wfe+message"
	amqpContentEncoding = "encoding/gob"
)

type amqpDelivery struct {
	amqp.Delivery
}

func (r *amqpDelivery) ID() string {
	return r.CorrelationId
}

func (r *amqpDelivery) Confirm() error {
	return r.Delivery.Ack(false)
}

func (r *amqpDelivery) Content(c interface{}) error {
	//un serialize the body and return a valid call
	decoder := gob.NewDecoder(bytes.NewBuffer(r.Body))
	if err := decoder.Decode(c); err != nil {
		return err
	}

	return nil
}

type amqpBroker struct {
	con *amqp.Connection
	//ctx     context.Context
	invalid bool
}

type amqpDispatcher struct {
	o  *RouteOptions
	ch *amqp.Channel
}

type amqpConsumer struct {
	o  *RouteOptions
	ch *amqp.Channel
}

func init() {
	RegisterBroker("amqp", func(u *url.URL) (Broker, error) {
		return NewAMQPBroker(u.String(), nil)
	})
}

func NewAMQPBroker(url string, Dial func(network, addr string) (net.Conn, error)) (Broker, error) {
	var broker amqpBroker
	if err := broker.init(url, Dial); err != nil {
		return nil, err
	}
	return &broker, nil
}

func (b *amqpBroker) init(url string, Dial func(network, addr string) (net.Conn, error)) error {
	c, err := amqp.DialConfig(url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Dial:      Dial,
	})

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

func (b *amqpBroker) makeRoute(o *RouteOptions) (*amqp.Channel, error) {
	//TODO: broker should remember the route options and reuse objects accordingly
	ch, err := b.con.Channel()
	if err != nil {
		return nil, err
	}

	if o != nil {
		if _, err := ch.QueueDeclare(o.Queue, o.Durable, o.AutoDelete, o.Exclusive, false, nil); err != nil {
			return nil, err
		}
	}

	return ch, err
}

func (b *amqpBroker) Consumer(o *RouteOptions) (Consumer, error) {
	ch, err := b.makeRoute(o)
	if err != nil {
		return nil, err
	}
	return &amqpConsumer{
		o:  o,
		ch: ch,
	}, nil
}

func (b *amqpBroker) Dispatcher() (Dispatcher, error) {
	ch, err := b.con.Channel()
	if err != nil {
		return nil, err
	}
	return &amqpDispatcher{
		ch: ch,
	}, nil
}

func (b *amqpBroker) Close() error {
	return b.con.Close()
}

func (b *amqpDispatcher) makeRoute(o *RouteOptions) error {
	if _, err := b.ch.QueueDeclare(o.Queue, o.Durable, o.AutoDelete, o.Exclusive, false, nil); err != nil {
		return err
	}

	return nil
}

func (b *amqpDispatcher) Dispatch(o *RouteOptions, msg *Message) (string, error) {
	b.makeRoute(o)

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(msg.Content); err != nil {
		return "", err
	}

	queue := o.Queue

	if queue == "" {
		return "", fmt.Errorf("queue is not set")
	}

	id := uuid.New()

	return id, b.ch.Publish("", queue, false, false, amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		ContentType:     amqpContentType,
		ContentEncoding: amqpContentEncoding,
		Body:            buffer.Bytes(),
		CorrelationId:   id,
	})
}

func (b *amqpDispatcher) Close() error {
	return b.ch.Close()
}

func (b *amqpConsumer) Consume() (<-chan Delivery, error) {
	msges, err := b.ch.Consume(b.o.Queue, "", b.o.AutoConfirm, b.o.Exclusive, false, false, nil)
	if err != nil {
		return nil, err
	}
	feeder := make(chan Delivery)

	go func() {
		defer close(feeder)
		for msg := range msges {
			if msg.ContentType != amqpContentType {
				log.Warningf("received a message with wrong content type '%s', ignoring.", msg.ContentType)
				break
			}

			feeder <- &amqpDelivery{
				Delivery: msg,
			}
		}
	}()

	return feeder, nil
}

func (b *amqpConsumer) Close() error {
	return b.ch.Close()
}
