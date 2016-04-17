package wfe

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

const (
	ResultQueueTmpl = "wfe.result.%s"
	DefaultTimeout  = -1
)

var (
	ErrTimeout = errors.New("timeout")
)

type ResultStore interface {
	Set(response *Response) error
	Get(id string, timeout int) (*Response, error)
}

type redisStore struct {
	pool         *redis.Pool
	getTimeout   int
	storeTimeout int
}

func NewRedisStore(server string, password string, getTimeout int, storeTimeout int) ResultStore {
	return &redisStore{
		getTimeout:   getTimeout,
		storeTimeout: storeTimeout,
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server)
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
	queue := fmt.Sprintf(ResultQueueTmpl, response.UUID)
	conn.Send("MULTI")
	conn.Send("LPUSH", queue, buffer.String())
	conn.Send("EXPIRE", queue, s.storeTimeout)
	_, err := conn.Do("EXEC")
	return err
}

func (s *redisStore) Get(uuid string, timeout int) (*Response, error) {
	conn := s.pool.Get()
	defer conn.Close()
	if timeout == DefaultTimeout {
		timeout = s.getTimeout
	}

	queue := fmt.Sprintf(ResultQueueTmpl, uuid)
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
