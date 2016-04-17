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
		store := newRedisStore(u.Host, pass, timeout, keep)
		return store, nil
	})
}
func newRedisStore(server string, password string, timeout int, keep int) ResultStore {
	return &redisStore{
		timeout: timeout,
		keep:    keep,
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
