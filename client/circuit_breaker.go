package client

import (
	"time"

	"github.com/savier89/circuitbreaker"
)

// CircuitBreakerConfig — конфигурация Circuit Breaker
type CircuitBreakerConfig struct {
	Name        string
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	IsError     func(error) bool
}

// NewCircuitBreaker — создаёт новый Circuit Breaker
func NewCircuitBreaker(cfg CircuitBreakerConfig) *circuitbreaker.CircuitBreaker {
	return circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		IsError:     cfg.IsError,
	})
}
