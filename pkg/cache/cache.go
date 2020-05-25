package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// Model is the struct we use for connecting and manipulating
// our Redis cache.
type Model struct {
	Conn redis.Conn
}

// Set sets a row by key in the Redis cache.
func (m *Model) Set(key, val string) error {
	_, err := m.Conn.Do("SET", key, val)
	if err != nil {
		return err
	}

	_, err = m.Conn.Do("SET", "time:"+key, time.Now().Unix())
	if err != nil {
		return err
	}

	return nil
}

// Get returns the value at a key, along with the time
// that that key was set.
func (m *Model) Get(key string) (string, int, error) {
	val, err := redis.String(m.Conn.Do("GET", key))
	if err != nil {
		return "", 0, err
	}

	time, err := redis.Int(m.Conn.Do("GET", "time:"+key))
	if err != nil {
		return "", 0, err
	}

	return val, time, nil
}
