package pulsar

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ishaqcherry9/depend/pkg/logger"
	workerpool "github.com/ishaqcherry9/depend/pkg/workpool"
	"go.uber.org/zap"

	"github.com/apache/pulsar-client-go/pulsar"
)

type consumer struct {
	internal pulsar.Consumer
	mu       sync.RWMutex
	closed   bool
	size     int
}

func (c *consumer) Receive(ctx context.Context) (*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("consumer is closed")
	}

	msg, err := c.internal.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	return &Message{
		Payload:  msg.Payload(),
		Key:      msg.Key(),
		Topic:    msg.Topic(),
		ID:       msg.ID(),
		internal: msg,
	}, nil
}

func (c *consumer) GetWorkerSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

func (c *consumer) SetWorkerSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.size = size
}

func (c *consumer) Consume(ctx context.Context, handleFn HandleMessageFn) error {
	if c.GetWorkerSize() < 1 {
		return fmt.Errorf("invalid worker size: %d", c.GetWorkerSize())
	}

	pool := workerpool.NewWorkerFIFOPool(c.GetWorkerSize(), func(err interface{}) {
		logger.Error(ctx, "message panic", zap.Any("panic", err))
	})
	defer workerpool.FIFOAntsRelease(pool)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case consumerMsg, ok := <-c.internal.Chan():
			if !ok {
				return errors.New("consumer channel closed")
			}
			err := workerpool.FIFOWorkerTaskSubmit(pool, func() {
				if err := handleFn(ctx, consumerMsg.Message); err != nil {
					c.internal.Nack(consumerMsg.Message)
				} else {
					c.internal.Ack(consumerMsg.Message)
				}
			})

			if err != nil {
				c.internal.Nack(consumerMsg.Message)
				logger.Error(ctx, "consumer task submit failed", zap.Error(err))
			}
		}
	}
}

func (c *consumer) Ack(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("invalid message")
	}

	return c.internal.Ack(msg.internal)
}

func (c *consumer) Nack(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("invalid message")
	}

	c.internal.Nack(msg.internal)
	return nil
}

func (c *consumer) AckCumulative(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("invalid message")
	}

	return c.internal.AckCumulative(msg.internal)
}

func (c *consumer) AckWithTransaction(txn pulsar.Transaction, msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	if txn == nil {
		return fmt.Errorf("transaction cannot be null")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("invalid message")
	}

	return c.internal.AckWithTxn(msg.internal, txn)
}

func (c *consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.internal.Close()
	return nil
}
