package wfe

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"net/url"
	"time"
)

const (
	resultQueueTmpl = "wfe.result.%s"

	//DefaultTimeout notates that a store should use it's default timeout
	DefaultTimeout = -1
)

var (
	//ErrTimeout timeout
	ErrTimeout = errors.New("timeout")
)

//ResultStore interface
type ResultStore interface {
	//Set a response in the result store
	Set(response *Response) error

	//Get a response from the result store, it blocks until a response is available or timeout is reached.
	//if timeout=DefaultTimeout, then the timeout is the default store timeout
	Get(id string, timeout int) (*Response, error)
}

type redisStore struct {
	pool    *redis.Pool
	timeout int
	keep    int
}

func init() {
	RegisterResultStore("redis", func(u *url.URL) (ResultStore, error) {
		timeout, err := parseInt(u.Query().Get("timeout"), 30)
		if err != nil {
			return nil, err
		}
		keep, err := parseInt(u.Query().Get("keep"), 3600)
		if err != nil {
			return nil, err
		}
		var pass string
		if u.User != nil {
			pass = u.User.Username()
		}
		store := NewRedisStore(u.Host, pass, timeout, keep)
		return store, nil
	})

	RegisterResultStore("discard", func(u *url.URL) (ResultStore, error) {
		return (*discardStore)(nil), nil
	})
}

func NewRedisStore(server string, password string, timeout int, keep int, options ...redis.DialOption) ResultStore {
	return &redisStore{
		timeout: timeout,
		keep:    keep,
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server, options...)
				if err != nil {
					return nil, err
				}

				if password != "" {
					if _, err := c.Do("AUTH", password); err != nil {
						c.Close()
						return nil, err
					}
				}

				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (s *redisStore) Set(response *Response) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(response); err != nil {
		return err
	}

	conn := s.pool.Get()
	defer conn.Close()
	queue := fmt.Sprintf(resultQueueTmpl, response.UUID)
	conn.Send("MULTI")
	conn.Send("LPUSH", queue, buffer.String())
	conn.Send("EXPIRE", queue, s.keep)
	_, err := conn.Do("EXEC")
	return err
}

func (s *redisStore) Get(uuid string, timeout int) (*Response, error) {
	conn := s.pool.Get()
	defer conn.Close()
	if timeout == DefaultTimeout {
		timeout = s.timeout
	}

	queue := fmt.Sprintf(resultQueueTmpl, uuid)
	result, err := redis.Bytes(conn.Do("BRPOPLPUSH", queue, queue, timeout))
	if err == redis.ErrNil {
		return nil, ErrTimeout
	} else if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(result)
	dec := gob.NewDecoder(buffer)
	var response Response
	if err := dec.Decode(&response); err != nil {
		return nil, err
	}

	return &response, err
}

type discardStore struct{}

func (s *discardStore) Set(response *Response) error {
	return nil
}

func (s *discardStore) Get(id string, timeout int) (*Response, error) {
	return nil, fmt.Errorf("Result has been discarded")
}
