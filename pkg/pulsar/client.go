package pulsar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"go.uber.org/zap"
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
			logger.Fatal(context.Background(), "pulsar client init failed", zap.Error(err))
		}
		gClient = &client{internal: internal}
	})
}

func GetClient() Client {
	if gClient == nil {
		logger.Fatal(context.Background(), "pulsar client get failed")
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

func (c *client) Producer(opts pulsar.ProducerOptions) (Producer, error) {
	internal, err := c.internal.CreateProducer(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &producer{internal: internal}, nil
}

func (c *client) Consumer(opts pulsar.ConsumerOptions) (Consumer, error) {
	internal, err := c.internal.Subscribe(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &consumer{internal: internal}, nil
}

func (c *client) MultiTopicConsumer(opts pulsar.ConsumerOptions) (Consumer, error) {
	internal, err := c.internal.Subscribe(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-topic consumer: %w", err)
	}

	return &consumer{internal: internal}, nil
}

func (c *client) Reader(opts pulsar.ReaderOptions) (Reader, error) {

	internal, err := c.internal.CreateReader(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reader: %w", err)
	}

	return &reader{internal: internal}, nil
}

func (c *client) BeginTransaction(timeout time.Duration) (pulsar.Transaction, error) {
	txn, err := c.internal.NewTransaction(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return txn, nil
}

func (c *client) Close() error {
	c.internal.Close()
	return nil
}
