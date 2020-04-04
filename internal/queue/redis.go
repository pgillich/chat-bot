// Package queue uses Redis
package queue

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/pgillich/chat-bot/internal/logger"
)

// ReceiveOnce calls PubSubConn.Receive() and unsubscribes
type ReceiveOnce func() ([]byte, error)

// RedisPublisher can be real or fake publisher
type RedisPublisher interface {
	Connect() error
	Close()
	Request(message []byte) error
}

// RealRedisPublisher is a real publisher
type RealRedisPublisher struct {
	Host           string
	Key            string
	User           string
	RequestChannel string

	pool *redis.Pool
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Connect connects to Redis
func (publisher *RealRedisPublisher) Connect() error {
	var err error

	publisher.pool = newPool(publisher.Host)

	conn := publisher.pool.Get()
	defer conn.Close() // nolint:errcheck

	_, err = conn.Do("SET", publisher.Key, publisher.User, "NX")
	if err != nil {
		return err
	}

	_, err = conn.Do("SADD", "users", publisher.User)
	if err != nil {
		return err
	}

	return nil
}

// Close deletes key+user and closes Redis sendConn
func (publisher *RealRedisPublisher) Close() {
	logger.Get().Info("Closing...")

	if publisher.pool != nil {
		conn := publisher.pool.Get()
		defer conn.Close() // nolint:errcheck

		conn.Do("DEL", publisher.Key)            // nolint:errcheck,gosec
		conn.Do("SREM", "users", publisher.User) // nolint:errcheck,gosec
		publisher.pool.Close()                   // nolint:errcheck,gosec

		publisher.pool = nil
	}
}

// Request sends a message to the queue
func (publisher *RealRedisPublisher) Request(message []byte) error {
	conn := publisher.pool.Get()
	defer conn.Close() // nolint:errcheck

	if _, err := conn.Do("PUBLISH", publisher.RequestChannel, message); err != nil {
		return err
	}

	return nil
}

// RedisSubscriber can be real or fake subscriber
type RedisSubscriber interface {
	Connect() error
	Close()
	Receive() interface{}
}

// RealRedisSubscriber is a real subscriber
type RealRedisSubscriber struct {
	Host           string
	Key            string
	User           string
	RequestChannel string

	conn        redis.Conn
	requestsPsc *redis.PubSubConn

	mx *sync.Mutex
}

// Connect connects to Redis
func (subscriber *RealRedisSubscriber) Connect() error {
	var err error

	subscriber.conn, err = redis.Dial("tcp", subscriber.Host)
	if err != nil {
		return err
	}

	subscriber.requestsPsc = &redis.PubSubConn{Conn: subscriber.conn}
	if err := subscriber.requestsPsc.Subscribe(subscriber.RequestChannel); err != nil {
		return err
	}

	subscriber.mx = new(sync.Mutex)

	return nil
}

// Close deletes key+user and closes Redis recvConn
func (subscriber *RealRedisSubscriber) Close() {
	subscriber.mx.Lock()
	defer subscriber.mx.Unlock()

	if subscriber.conn != nil {
		logger.Get().Info("Closing...")

		if err := subscriber.requestsPsc.Unsubscribe(subscriber.RequestChannel); err != nil {
			logger.Get().Warning("cannot unsubscribe request channel", err)
		}

		subscriber.conn.Close() // nolint:errcheck,gosec

		subscriber.conn = nil
	}
}

// Receive listens to request channel
func (subscriber *RealRedisSubscriber) Receive() interface{} {
	return subscriber.requestsPsc.Receive()
}
