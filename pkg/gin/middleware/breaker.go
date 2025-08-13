package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ishaqcherry9/depend/pkg/container/group"
	"github.com/ishaqcherry9/depend/pkg/gin/response"
	"github.com/ishaqcherry9/depend/pkg/shield/circuitbreaker"
)

var ErrNotAllowed = circuitbreaker.ErrNotAllowed

type CircuitBreakerOption func(*circuitBreakerOptions)

type circuitBreakerOptions struct {
	group          *group.Group
	validCodes     map[int]struct{}
	degradeHandler func(c *gin.Context)
}

func defaultCircuitBreakerOptions() *circuitBreakerOptions {
	return &circuitBreakerOptions{
		group: group.NewGroup(func() interface{} {
			return circuitbreaker.NewBreaker()
		}),
		validCodes: map[int]struct{}{
			http.StatusInternalServerError: {},
			http.StatusServiceUnavailable:  {},
		},
	}
}

func (o *circuitBreakerOptions) apply(opts ...CircuitBreakerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithGroup(g *group.Group) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		if g != nil {
			o.group = g
		}
	}
}

func WithValidCode(code ...int) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		for _, c := range code {
			o.validCodes[c] = struct{}{}
		}
	}
}

func WithDegradeHandler(handler func(c *gin.Context)) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		o.degradeHandler = handler
	}
}

func CircuitBreaker(opts ...CircuitBreakerOption) gin.HandlerFunc {
	o := defaultCircuitBreakerOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		breaker := o.group.Get(c.FullPath()).(circuitbreaker.CircuitBreaker)
		if err := breaker.Allow(); err != nil {
			breaker.MarkFailed()
			if o.degradeHandler != nil {
				o.degradeHandler(c)
			} else {
				response.Output(c, http.StatusServiceUnavailable, err.Error())
			}
			c.Abort()
			return
		}

		c.Next()

		code := c.Writer.Status()
		_, isHit := o.validCodes[code]
		if isHit {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
	}
}
