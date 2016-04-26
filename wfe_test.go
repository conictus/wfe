package wfe

import (
	"github.com/stretchr/testify/mock"
	"net/url"
)

type TestBroker struct {
	mock.Mock
}

type TestDispatcher struct {
	mock.Mock
}

type TestConsumer struct {
	mock.Mock
}

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

func init() {
	RegisterBroker("test", func(_ *url.URL) (Broker, error) {
		return &TestBroker{}, nil
	})
}

func (b *TestBroker) Close() error {
	args := b.Called()
	return args.Error(0)
}

func (b *TestBroker) Dispatcher(o *RouteOptions) (Dispatcher, error) {
	args := b.Called(o)
	return args.Get(0).(Dispatcher), args.Error(1)
}

func (b *TestBroker) Consumer(o *RouteOptions) (Consumer, error) {
	args := b.Called(o)
	return args.Get(0).(Consumer), args.Error(1)
}
