package pulsar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type reader struct {
	internal pulsar.Reader
	mu       sync.RWMutex
	closed   bool
}

func (r *reader) Next(ctx context.Context) (*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, fmt.Errorf("Reader已关闭")
	}

	msg, err := r.internal.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取消息失败: %w", err)
	}

	return &Message{
		Payload:  msg.Payload(),
		Key:      msg.Key(),
		Topic:    msg.Topic(),
		ID:       msg.ID(),
		internal: msg,
	}, nil
}

func (r *reader) HasNext() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return false
	}

	return r.internal.HasNext()
}

func (r *reader) Seek(messageID MessageID) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("Reader已关闭")
	}

	return r.internal.Seek(messageID)
}

func (r *reader) SeekByTime(timestamp time.Time) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("Reader已关闭")
	}

	return r.internal.SeekByTime(timestamp)
}

func (r *reader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	r.internal.Close()
	return nil
}
