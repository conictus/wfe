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

func (d *testDispatcher) Dispatch(m *Message) (string, error) {
	args := d.Called(m)
	return args.String(0), args.Error(1)
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
	v := args.Get(0)
	var r *Response
	if v != nil {
		r = v.(*Response)
	}

	return r, args.Error(1)
}

func NewTestContext(id string, client Client) *Context {
	return &Context{
		id:     id,
		client: client,
		values: make(map[string]interface{}),
	}
}

type TestClient struct {
	mock.Mock
}

func (tc *TestClient) Apply(req Request) (Result, error) {
	args := tc.Called(req)
	r := args.Get(0)
	if r == nil {
		return nil, args.Error(1)
	}

	return r.(Result), args.Error(1)
}

func (tc *TestClient) Group(requests ...Request) (GroupResult, error) {
	var in []interface{}
	for _, r := range requests {
		in = append(in, r)
	}

	args := tc.Called(in...)

	r := args.Get(0)
	if r == nil {
		return nil, args.Error(1)
	}

	return r.(GroupResult), args.Error(1)
}

func (tc *TestClient) Chain(request Request, chain ...PartialRequest) (Result, error) {
	in := []interface{}{request}
	for _, r := range chain {
		in = append(in, r)
	}

	args := tc.Called(in...)
	r := args.Get(0)
	if r == nil {
		return nil, args.Error(1)
	}

	return r.(Result), args.Error(1)
}

func (tc *TestClient) Chord(callback PartialRequest, requests ...Request) (Result, error) {
	in := []interface{}{callback}
	for _, r := range requests {
		in = append(in, r)
	}

	args := tc.Called(in...)
	r := args.Get(0)
	if r == nil {
		return nil, args.Error(1)
	}

	return r.(Result), args.Error(1)
}

func (tc *TestClient) ResultFor(id string) Result {
	args := tc.Called(id)
	r := args.Get(0)
	if r == nil {
		return nil
	}

	return r.(Result)
}

func (tc *TestClient) Close() error {
	args := tc.Called()
	return args.Error(0)
}

type DummyClient struct {
}

func (dc *DummyClient) Apply(req Request) (Result, error) {
	return nil, nil
}

func (dc *DummyClient) Group(requests ...Request) (GroupResult, error) {
	return nil, nil
}

func (dc *DummyClient) Chain(request Request, chain ...PartialRequest) (Result, error) {
	return nil, nil
}

func (dc *DummyClient) Chord(callback PartialRequest, requests ...Request) (Result, error) {
	return nil, nil
}

func (dc *DummyClient) ResultFor(id string) Result {
	return nil
}

func (dc *DummyClient) Close() error {
	return nil
}
