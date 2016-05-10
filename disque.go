package wfe

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/goware/disque"
	"net/url"
	"time"
)

func init() {
	RegisterBroker("disque", func(u *url.URL) (Broker, error) {
		pool, err := disque.New(u.Host)
		if err != nil {
			return nil, err
		}

		if err := pool.Ping(); err != nil {
			return nil, err
		}

		retry, err := parseInt(u.Query().Get("retry"), 0)
		if err != nil {
			return nil, err
		}

		return &disqueBroker{pool: pool.RetryAfter(time.Duration(retry) * time.Second)}, nil
	})
}

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

func (b *disqueBroker) Close() error {
	return b.pool.Close()
}

func (b *disqueBroker) Dispatcher(o *RouteOptions) (Dispatcher, error) {
	return &disqueBroker{
		pool: b.pool,
		opt:  o,
	}, nil
}

func (b *disqueBroker) Consumer(o *RouteOptions) (Consumer, error) {
	return &disqueBroker{
		pool: b.pool,
		opt:  o,
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
