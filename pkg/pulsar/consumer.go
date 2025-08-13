package pulsar

import (
	"context"
	"errors"
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/logger"
	workerpool "github.com/ishaqcherry9/depend/pkg/workpool"
	"go.uber.org/zap"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
)

type consumer struct {
	internal pulsar.Consumer
	mu       sync.RWMutex
	closed   bool
}

func (c *consumer) Receive(ctx context.Context) (*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("消费者已关闭")
	}

	msg, err := c.internal.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("接收消息失败: %w", err)
	}

	return &Message{
		Payload:  msg.Payload(),
		Key:      msg.Key(),
		Topic:    msg.Topic(),
		ID:       msg.ID(),
		internal: msg,
	}, nil
}

func (c *consumer) Consume(ctx context.Context, handleFn HandleMessageFn) error {
	pool := workerpool.NewWorkerFIFOPool(10, func(err interface{}) {
		logger.Error("message panic", zap.Any("panic", err))
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
				logger.Error("consumer task submit failed", zap.Error(err))
			}
		}
	}
}

func (c *consumer) Ack(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("消费者已关闭")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("无效消息")
	}

	return c.internal.Ack(msg.internal)
}

func (c *consumer) Nack(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("消费者已关闭")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("无效消息")
	}

	c.internal.Nack(msg.internal)
	return nil
}

func (c *consumer) AckCumulative(msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("消费者已关闭")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("无效消息")
	}

	return c.internal.AckCumulative(msg.internal)
}

func (c *consumer) AckWithTransaction(txn pulsar.Transaction, msg *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("消费者已关闭")
	}

	if txn == nil {
		return fmt.Errorf("事务不能为空")
	}

	if msg == nil || msg.internal == nil {
		return fmt.Errorf("无效消息")
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
