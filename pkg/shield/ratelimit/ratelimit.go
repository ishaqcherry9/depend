package ratelimit

import (
	"errors"
)

var (
	ErrLimitExceed = errors.New("rate limit exceeded")
)

type DoneFunc func(DoneInfo)

type DoneInfo struct {
	Err error
}

type Limiter interface {
	Allow() (DoneFunc, error)
}
