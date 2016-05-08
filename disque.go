package wfe

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/goware/disque"
)

////Broker interface
//type Broker interface {
//	//Close broker connections. Close must force the broker Consumer to return
//	Close() error
//
//	//Dispatcher gets a dispatcher instance according to the RouterOptions specified.
//	Dispatcher(o *RouteOptions) (Dispatcher, error)
//
//	//Consumer gets a consumer instance according to the RouterOptions specified.
//	Consumer(o *RouteOptions) (Consumer, error)
//}

type disqueBroker struct {
	pool *disque.Pool
	opt  *RouteOptions
}

type disqueDelivery struct {
	pool *disque.Pool
	j    *disque.Job
}

func (d *disqueDelivery) ID() string {
	return d.j.ID
}

func (d *disqueDelivery) Confirm() error {
	return d.pool.Ack(d.j)
}

func (d *disqueDelivery) Content(c interface{}) error {
	//un serialize the body and return a valid call
	decoder := gob.NewDecoder(bytes.NewBufferString(d.j.Data))
	if err := decoder.Decode(c); err != nil {
		return err
	}

	return nil
}

//type disqueDispatcher struct {
//	pool *disque.Pool
//}
//
//type disqueConsumer struct {
//	pool disque.Pool
//}

func (b *disqueBroker) Close() error {
	return b.pool.Close()
}

func (b *disqueBroker) Dispatcher(o *RouteOptions) (Dispatcher, error) {
	return &disqueBroker{
		pool: b.pool,
		opt:  b.opt,
	}, nil
}

func (b *disqueBroker) Consumer(o *RouteOptions) (Consumer, error) {
	return &disqueBroker{
		pool: b.pool,
		opt:  b.opt,
	}, nil
}

func (b *disqueBroker) Dispatch(msg *Message) (string, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(msg.Content); err != nil {
		return "", err
	}

	queue := msg.Queue
	if b.opt != nil {
		queue = b.opt.Queue
	}

	if queue == "" {
		return "", fmt.Errorf("queue is not set")
	}

	job, err := b.pool.Add(buffer.String(), queue)
	return job.ID, err
}

func (b *disqueBroker) Consume() (<-chan Delivery, error) {
	deliveries := make(chan Delivery)
	go func() {
		defer close(deliveries)
		for {
			job, err := b.pool.Get(b.opt.Queue)
			if err != nil {
				return
			}
			if b.opt.AutoConfirm {
				b.pool.Ack(job)
			}

			deliveries <- &disqueDelivery{
				pool: b.pool,
				j:    job,
			}
		}
	}()

	return deliveries, nil
}

//
////Dispatcher interface
//type Dispatcher interface {
//	//Close dispatcher
//	Close() error
//
//	//Dispatch msg
//	Dispatch(msg *Message) error
//}
//
////Consumer interfaec
//type Consumer interface {
//	//Close consumer
//	Close() error
//
//	//Consume gets a Deliver channel. the delivery channel is auto closed if the broker connection is lost
//	//or the consumer is closed explicitly.
//	Consume() (<-chan Delivery, error)
//}
