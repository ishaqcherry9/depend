package pulsar

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	gClient Client
	once    sync.Once
)

type client struct {
	internal pulsar.Client
}

func InitClient(config *ClientConfig) {
	if config == nil {
		return
	}
	once.Do(func() {

		setDefaults(config)
		options := pulsar.ClientOptions{
			URL: config.ServiceURL,
		}

		if config.Auth != nil {
			switch config.Auth.Type {
			case "token":
				options.Authentication = pulsar.NewAuthenticationToken(config.Auth.Token)
			case "tls":
				options.Authentication = pulsar.NewAuthenticationTLS(config.Auth.Cert, config.Auth.Key)
			}
		}

		internal, err := pulsar.NewClient(options)
		if err != nil {
			logger.Fatal("pulsar client init failed", zap.Error(err))
		}
		gClient = &client{internal: internal}
	})
}

func GetClient() Client {
	if gClient == nil {
		logger.Fatal("pulsar client get failed")
	}
	return gClient
}

func CloseClient() error {
	if gClient != nil {
		if err := gClient.Close(); err != nil {
			return err
		}
		gClient = nil
	}
	return nil
}

func (c *client) Producer(topic string, opts ...ProducerOption) (Producer, error) {
	options := pulsar.ProducerOptions{
		Topic: topic,
	}

	for _, opt := range opts {
		opt(&options)
	}

	internal, err := c.internal.CreateProducer(options)
	if err != nil {
		return nil, fmt.Errorf("创建生产者失败: %w", err)
	}

	return &producer{internal: internal}, nil
}

func (c *client) Consumer(topic, subscription string, opts ...ConsumerOption) (Consumer, error) {
	options := pulsar.ConsumerOptions{
		Topic:                          topic,
		SubscriptionName:               subscription,
		EnableBatchIndexAcknowledgment: true,
	}

	for _, opt := range opts {
		opt(&options)
	}

	internal, err := c.internal.Subscribe(options)
	if err != nil {
		return nil, fmt.Errorf("创建消费者失败: %w", err)
	}

	return &consumer{internal: internal}, nil
}

func (c *client) MultiTopicConsumer(topicPattern, subscription string, opts ...ConsumerOption) (Consumer, error) {
	options := pulsar.ConsumerOptions{
		TopicsPattern:    topicPattern,
		SubscriptionName: subscription,
	}

	for _, opt := range opts {
		opt(&options)
	}

	internal, err := c.internal.Subscribe(options)
	if err != nil {
		return nil, fmt.Errorf("创建多主题消费者失败: %w", err)
	}

	return &consumer{internal: internal}, nil
}

func (c *client) Reader(topic string, startMessageID MessageID, opts ...ReaderOption) (Reader, error) {
	options := pulsar.ReaderOptions{
		Topic:          topic,
		StartMessageID: startMessageID,
	}

	for _, opt := range opts {
		opt(&options)
	}

	internal, err := c.internal.CreateReader(options)
	if err != nil {
		return nil, fmt.Errorf("创建Reader失败: %w", err)
	}

	return &reader{internal: internal}, nil
}

func (c *client) BeginTransaction(timeout time.Duration) (pulsar.Transaction, error) {
	txn, err := c.internal.NewTransaction(timeout)
	if err != nil {
		return nil, fmt.Errorf("创建事务失败: %w", err)
	}

	return txn, nil
}

func (c *client) Close() error {
	c.internal.Close()
	return nil
}
