package core

import "context"

// Message is the envelope passed through messaging brokers.
type Message struct {
	Topic   string
	Key     []byte
	Payload []byte
	Headers map[string]string
}

// MessageHandler is the function signature for consuming messages.
type MessageHandler func(ctx context.Context, msg Message) error

// Publisher is the contract for publishing messages (e.g. ss-keel-amqp, ss-keel-kafka).
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

// Subscriber is the contract for consuming messages from a topic.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}
