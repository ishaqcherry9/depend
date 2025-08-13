package pulsar

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type HandleMessageFn func(ctx context.Context, msg pulsar.Message) error

type BusinessMessage struct {
	ID      string
	Payload []byte
}

type BatchResult struct {
	SuccessIDs         []MessageID
	FailureBusinessIDs []string
}

type Client interface {
	Producer(topic string, opts ...ProducerOption) (Producer, error)

	Consumer(topic, subscription string, opts ...ConsumerOption) (Consumer, error)

	MultiTopicConsumer(topicPattern, subscription string, opts ...ConsumerOption) (Consumer, error)

	Reader(topic string, startMessageID MessageID, opts ...ReaderOption) (Reader, error)

	BeginTransaction(timeout time.Duration) (pulsar.Transaction, error)

	Close() error
}

type Producer interface {
	Send(ctx context.Context, payload []byte) error
	SendWithKey(ctx context.Context, key string, payload []byte) error

	SendDelayed(ctx context.Context, payload []byte, delay time.Duration) error

	SendBatch(ctx context.Context, messages []BusinessMessage) (*BatchResult, error)

	SendWithTransaction(ctx context.Context, txn pulsar.Transaction, payload []byte) error

	Close() error
}

type Consumer interface {
	Receive(ctx context.Context) (*Message, error)
	Consume(ctx context.Context, handleFn HandleMessageFn) error
	Ack(msg *Message) error
	Nack(msg *Message) error
	AckCumulative(msg *Message) error
	AckWithTransaction(txn pulsar.Transaction, msg *Message) error
	Close() error
}

type Reader interface {
	Next(ctx context.Context) (*Message, error)
	HasNext() bool
	Seek(messageID MessageID) error
	SeekByTime(timestamp time.Time) error
	Close() error
}

type Message struct {
	Payload  []byte
	Key      string
	Topic    string
	ID       MessageID
	internal pulsar.Message
}

func (m *Message) Internal() pulsar.Message {
	return m.internal
}

type MessageID = pulsar.MessageID

func EarliestMessageID() MessageID {
	return pulsar.EarliestMessageID()
}

func LatestMessageID() MessageID {
	return pulsar.LatestMessageID()
}
