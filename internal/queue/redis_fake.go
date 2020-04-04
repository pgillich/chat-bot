package queue

import (
	"github.com/garyburd/redigo/redis"
)

const fakeRedisQueueSize = 10

// FakeRedis is a fake implementation for RedisPublisher and RedisSubscriber
// Only for one-to-one connection (1 publisher, 1 subscriber)
// TODO more subscriber, more channel
type FakeRedis struct {
	queue chan redis.Message
}

// Connect makes a new queue
// It's called twice: on publisher and on substriber side
func (fakeRedis *FakeRedis) Connect() error {
	fakeRedis.queue = make(chan redis.Message, fakeRedisQueueSize)

	return nil
}

// Close does nothing
// It's called twice: on publisher and on substriber side
// TODO empty queue
func (fakeRedis *FakeRedis) Close() {
}

// Request puts a message into queue
// TODO timeout (error) if queue is full
func (fakeRedis *FakeRedis) Request(message []byte) error {
	fakeRedis.queue <- redis.Message{Data: message}

	return nil
}

// Receive reads (waits for) a message from the queue
func (fakeRedis *FakeRedis) Receive() interface{} {
	message := <-fakeRedis.queue
	return message
}
