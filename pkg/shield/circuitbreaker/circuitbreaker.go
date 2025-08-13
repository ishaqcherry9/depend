package circuitbreaker

import (
	"errors"
)

var ErrNotAllowed = errors.New("circuitbreaker: not allowed for circuit open")

type CircuitBreaker interface {
	Allow() error
	MarkSuccess()
	MarkFailed()
}
