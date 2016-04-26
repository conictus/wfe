package wfe

import (
	"github.com/stretchr/testify/mock"
)

type testBroker struct {
	mock.Mock
}

func (b *testBroker) Close() error {
	args := b.Called()
	return args.Error(0)
}

func (b *testBroker) Dispatcher(o *RouteOptions) (Dispatcher, error) {
	args := b.Called(o)
	return args.Get(0).(Dispatcher), args.Error(1)
}

func (b *testBroker) Consumer(o *RouteOptions) (Consumer, error) {
	args := b.Called(o)
	return args.Get(0).(Consumer), args.Error(1)
}

type testDispatcher struct {
	mock.Mock
}

func (d *testDispatcher) Close() error {
	args := d.Called()
	return args.Error(0)
}

func (d *testDispatcher) Dispatch(m *Message) error {
	args := d.Called(m)
	return args.Error(0)
}

type testConsumer struct {
	mock.Mock
}

func (d *testConsumer) Close() error {
	args := d.Called()
	return args.Error(0)
}

func (d *testConsumer) Consume() (<-chan Delivery, error) {
	args := d.Called()
	return args.Get(0).(<-chan Delivery), args.Error(1)
}

type testStore struct {
	mock.Mock
}

func (t *testStore) Set(response *Response) error {
	args := t.Called(response)
	return args.Error(0)
}

func (t *testStore) Get(id string, timeout int) (*Response, error) {
	args := t.Called(id, timeout)
	return args.Get(0).(*Response), args.Error(1)
}
