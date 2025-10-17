package pulsar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type producer struct {
	internal pulsar.Producer
	mu       sync.RWMutex
	closed   bool
}

func (p *producer) Send(ctx context.Context, payload []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	_, err := p.internal.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *producer) SendWithKey(ctx context.Context, key string, payload []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	_, err := p.internal.Send(ctx, &pulsar.ProducerMessage{
		Key:     key,
		Payload: payload,
	})

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *producer) SendDelayed(ctx context.Context, payload []byte, delay time.Duration) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	_, err := p.internal.Send(ctx, &pulsar.ProducerMessage{
		Payload:      payload,
		DeliverAfter: delay,
	})

	if err != nil {
		return fmt.Errorf("failed to send delayed message: %w", err)
	}

	return nil
}

func (p *producer) SendBatch(ctx context.Context, messages []BusinessMessage) (*BatchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("producer is closed")
	}

	if len(messages) <= 0 {
		return &BatchResult{}, nil
	}

	result := &BatchResult{
		SuccessIDs:         make([]MessageID, 0, len(messages)),
		FailureBusinessIDs: make([]string, 0),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, msg := range messages {
		wg.Add(1)
		businessMsg := msg
		p.internal.SendAsync(ctx, &pulsar.ProducerMessage{
			Payload: businessMsg.Payload,
		}, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.FailureBusinessIDs = append(result.FailureBusinessIDs, businessMsg.ID)
			} else {
				result.SuccessIDs = append(result.SuccessIDs, id)
			}
		})
	}
	wg.Wait()

	if err := p.internal.FlushWithCtx(ctx); err != nil {
		return result, fmt.Errorf("failed to flush messages: %w", err)
	}
	return result, nil
}

func (p *producer) SendAsync(ctx context.Context, msg *pulsar.ProducerMessage, callback func(id MessageID, msg *pulsar.ProducerMessage, err error)) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		if callback != nil {
			callback(nil, msg, fmt.Errorf("producer is closed"))
		}
		return
	}

	p.internal.SendAsync(ctx, msg, callback)
}
func (p *producer) SendWithTransaction(ctx context.Context, txn pulsar.Transaction, payload []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	if txn == nil {
		return fmt.Errorf("transaction cannot be null")
	}

	_, err := p.internal.Send(ctx, &pulsar.ProducerMessage{
		Payload:     payload,
		Transaction: txn,
	})

	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

func (p *producer) FlushWithCtx(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	return p.internal.FlushWithCtx(ctx)
}

func (p *producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	p.internal.Close()
	return nil
}
