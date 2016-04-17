package wfe

import (
	"net/url"
	"sync"
)

var (
	brokers = map[string]BrokerFactory{}
	stores  = map[string]ResultStoreFactory{}

	bm sync.Mutex
	sm sync.Mutex
)

type BrokerFactory func(u *url.URL) (Broker, error)
type ResultStoreFactory func(u *url.URL) (ResultStore, error)

type Options struct {
	Broker string
	Store  string
}

func RegisterBroker(scheme string, factory BrokerFactory) {
	bm.Lock()
	defer bm.Unlock()

	brokers[scheme] = factory
}

func RegisterResultStore(scheme string, factory ResultStoreFactory) {
	sm.Lock()
	defer sm.Unlock()

	stores[scheme] = factory
}

func (o *Options) GetBroker() (Broker, error) {
	u, err := url.Parse(o.Broker)
	if err != nil {
		log.Fatalf("failed to parse broker url: %s", o.Broker)
		return nil, err
	}

	bm.Lock()
	defer bm.Unlock()

	factory, ok := brokers[u.Scheme]
	if !ok {
		log.Fatalf("unknown broker %s", u.Scheme)
	}

	return factory(u)
}

func (o *Options) GetStore() (ResultStore, error) {
	u, err := url.Parse(o.Store)
	if err != nil {
		log.Fatalf("failed to parse broker url: %s", o.Broker)
		return nil, err
	}

	sm.Lock()
	defer sm.Unlock()

	factory, ok := stores[u.Scheme]
	if !ok {
		log.Fatalf("unknown store %s", u.Scheme)
	}

	return factory(u)
}
