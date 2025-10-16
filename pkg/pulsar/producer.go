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

	if len(messages) == 0 {
		return &BatchResult{}, nil
	}

	type sendResult struct {
		businessID string
		messageID  MessageID
		success    bool
	}

	resultCh := make(chan sendResult, len(messages))
	var wg sync.WaitGroup

	for _, msg := range messages {
		wg.Add(1)
		go func(businessMsg BusinessMessage) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				resultCh <- sendResult{
					businessID: businessMsg.ID,
					success:    false,
				}
				return
			default:
			}

			messageID, err := p.internal.Send(ctx, &pulsar.ProducerMessage{
				Payload: businessMsg.Payload,
			})

			if err != nil {
				resultCh <- sendResult{
					businessID: businessMsg.ID,
					success:    false,
				}
			} else {
				resultCh <- sendResult{
					businessID: businessMsg.ID,
					messageID:  messageID,
					success:    true,
				}
			}
		}(msg)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	result := &BatchResult{
		SuccessIDs:         make([]MessageID, 0),
		FailureBusinessIDs: make([]string, 0),
	}

	for res := range resultCh {
		if res.success {
			result.SuccessIDs = append(result.SuccessIDs, res.messageID)
		} else {
			result.FailureBusinessIDs = append(result.FailureBusinessIDs, res.businessID)
		}
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
