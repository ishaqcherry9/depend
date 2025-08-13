package pulsar

import (
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type ProducerOption func(options *pulsar.ProducerOptions)

type ConsumerOption func(options *pulsar.ConsumerOptions)

type ReaderOption func(options *pulsar.ReaderOptions)

func WithProducerName(name string) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.Name = name
	}
}

func WithSendTimeout(timeout time.Duration) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.SendTimeout = timeout
	}
}

func WithBatching(enabled bool) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.DisableBatching = !enabled
	}
}

func WithCompression(compressionType pulsar.CompressionType) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.CompressionType = compressionType
	}
}

func WithProperties(properties map[string]string) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.Properties = properties
	}
}

func WithSchema(schema pulsar.Schema) ProducerOption {
	return func(options *pulsar.ProducerOptions) {
		options.Schema = schema
	}
}

func WithConsumerName(name string) ConsumerOption {
	return func(options *pulsar.ConsumerOptions) {
		options.Name = name
	}
}

func WithSubscriptionType(subType pulsar.SubscriptionType) ConsumerOption {
	return func(options *pulsar.ConsumerOptions) {
		options.Type = subType
	}
}

func WithReceiverQueueSize(size int) ConsumerOption {
	return func(options *pulsar.ConsumerOptions) {
		options.ReceiverQueueSize = size
	}
}

func WithDLQ(maxDeliveries uint32, deadLetterTopic string) ConsumerOption {
	return func(options *pulsar.ConsumerOptions) {
		options.DLQ = &pulsar.DLQPolicy{
			MaxDeliveries:   maxDeliveries,
			DeadLetterTopic: deadLetterTopic,
		}
	}
}

func WithInitialPosition(position pulsar.SubscriptionInitialPosition) ConsumerOption {
	return func(options *pulsar.ConsumerOptions) {
		options.SubscriptionInitialPosition = position
	}
}

func WithReaderName(name string) ReaderOption {
	return func(options *pulsar.ReaderOptions) {
		options.Name = name
	}
}

func WithReaderStartInclusive(inclusive bool) ReaderOption {
	return func(options *pulsar.ReaderOptions) {
		options.StartMessageIDInclusive = inclusive
	}
}
