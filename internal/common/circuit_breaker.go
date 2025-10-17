package common

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	name            string
	maxFailures     uint
	resetTimeout    time.Duration
	halfOpenTimeout time.Duration
	state           CircuitState
	failures        uint
	lastFailureTime time.Time
	lastSuccessTime time.Time
	nextRetryTime   time.Time
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	Name            string
	MaxFailures     uint
	ResetTimeout    time.Duration
	HalfOpenTimeout time.Duration
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are made in half-open state
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout == 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.HalfOpenTimeout == 0 {
		config.HalfOpenTimeout = 30 * time.Second
	}

	return &CircuitBreaker{
		name:            config.Name,
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenTimeout: config.HalfOpenTimeout,
		state:           StateClosed,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn(ctx)

	cb.afterRequest(err)
	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if now.After(cb.nextRetryTime) {
			// Transition to half-open state
			cb.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		// Allow limited requests in half-open state
		if now.Before(cb.nextRetryTime) {
			return ErrTooManyRequests
		}
		return nil
	}

	return nil
}

// afterRequest updates the circuit breaker state based on the result
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	if err == nil {
		cb.onSuccess(now)
	} else {
		cb.onFailure(now)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess(now time.Time) {
	cb.lastSuccessTime = now

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		// Transition back to closed state after successful request
		cb.state = StateClosed
		cb.failures = 0
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure(now time.Time) {
	cb.failures++
	cb.lastFailureTime = now

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.nextRetryTime = now.Add(cb.resetTimeout)
		}
	case StateHalfOpen:
		// Immediately transition back to open state on failure
		cb.state = StateOpen
		cb.nextRetryTime = now.Add(cb.resetTimeout)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	stateStr := ""
	switch cb.state {
	case StateClosed:
		stateStr = "closed"
	case StateHalfOpen:
		stateStr = "half-open"
	case StateOpen:
		stateStr = "open"
	}

	return map[string]interface{}{
		"name":              cb.name,
		"state":             stateStr,
		"failures":          cb.failures,
		"last_failure_time": cb.lastFailureTime,
		"last_success_time": cb.lastSuccessTime,
		"next_retry_time":   cb.nextRetryTime,
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
}

// CircuitBreakerRegistry manages multiple circuit breakers
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

var globalRegistry = &CircuitBreakerRegistry{
	breakers: make(map[string]*CircuitBreaker),
}

// GetCircuitBreaker gets or creates a circuit breaker
func GetCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if cb, exists := globalRegistry.breakers[name]; exists {
		return cb
	}

	if config == nil {
		config = &CircuitBreakerConfig{
			Name:            name,
			MaxFailures:     5,
			ResetTimeout:    60 * time.Second,
			HalfOpenTimeout: 30 * time.Second,
		}
	}

	cb := NewCircuitBreaker(*config)
	globalRegistry.breakers[name] = cb
	return cb
}

// GetAllStats returns statistics for all circuit breakers
func GetAllCircuitBreakerStats() map[string]interface{} {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, cb := range globalRegistry.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}
