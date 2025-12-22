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

type MyConsumerOptions struct {
	Opts       pulsar.ConsumerOptions
	WorkerSize int
	MaxRetry   int
}

type Client interface {
	Producer(opts pulsar.ProducerOptions) (Producer, error)
	Consumer(opts pulsar.ConsumerOptions) (Consumer, error)
	//MyConsumer(opts MyConsumerOptions) (Consumer, error)
	MultiTopicConsumer(opts pulsar.ConsumerOptions) (Consumer, error)
	Reader(opts pulsar.ReaderOptions) (Reader, error)
	BeginTransaction(timeout time.Duration) (pulsar.Transaction, error)
	Close() error
}

type Producer interface {
	Send(ctx context.Context, payload []byte) error
	SendWithKey(ctx context.Context, key string, payload []byte) error
	SendDelayed(ctx context.Context, payload []byte, delay time.Duration) error
	SendBatch(ctx context.Context, messages []BusinessMessage) (*BatchResult, error)
	SendAsync(ctx context.Context, msg *pulsar.ProducerMessage, callback func(id MessageID, msg *pulsar.ProducerMessage, err error))
	SendWithTransaction(ctx context.Context, txn pulsar.Transaction, payload []byte) error
	FlushWithCtx(ctx context.Context) error
	Close() error
}

// Get|SetWorkerSize 暂时蹩脚放在interface中，可新建MyConsumerOptions结构体，包括如下参数，看后续业务诉求。
// pulsar.ConsumerOptions、WorkerSize、MaxRetry等，放在client创建时指定。
type Consumer interface {
	GetWorkerSize() int
	SetWorkerSize(size int)
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
