package domain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type sagaFaultKey struct{}

// SagaFaultConfig holds fault injection params populated from X-Saga-* request headers.
// All methods are nil-safe — nil config is a no-op.
// Active only when SAGA_TEST_MODE=true env var is set on bank-service.
type SagaFaultConfig struct {
	ForceFailStep string // step name (e.g. "RESERVE_FUNDS") to force-fail
	ForceFailKind string // "before"|"after" (default "before")
	CompFailStep  string // compensator step name to force-fail
	CompFailTimes int    // how many times compensator must fail before succeeding
	DelayStep     string // step name to inject delay at
	DelayMS       int    // delay in milliseconds

	mu       sync.Mutex
	attempts map[string]int // per-step compensator attempt counter
}

// NewSagaFaultConfig allocates a SagaFaultConfig with initialized attempt map.
func NewSagaFaultConfig() *SagaFaultConfig {
	return &SagaFaultConfig{attempts: make(map[string]int)}
}

// WithFaultConfig attaches cfg to ctx for propagation into the SAGA goroutine.
func WithFaultConfig(ctx context.Context, cfg *SagaFaultConfig) context.Context {
	return context.WithValue(ctx, sagaFaultKey{}, cfg)
}

// FaultConfigFromCtx retrieves the SagaFaultConfig from ctx; returns nil if absent.
func FaultConfigFromCtx(ctx context.Context) *SagaFaultConfig {
	v, _ := ctx.Value(sagaFaultKey{}).(*SagaFaultConfig)
	return v
}

// CheckStepBefore returns an injected error if a "before" fault is configured for step.
func (c *SagaFaultConfig) CheckStepBefore(step string) error {
	if c == nil || c.ForceFailStep != step || c.ForceFailKind == "after" {
		return nil
	}
	return fmt.Errorf("fault injection: step %s forced fail (before side effects)", step)
}

// CheckStepAfter returns an injected error if an "after" fault is configured for step.
func (c *SagaFaultConfig) CheckStepAfter(step string) error {
	if c == nil || c.ForceFailStep != step || c.ForceFailKind != "after" {
		return nil
	}
	return fmt.Errorf("fault injection: step %s forced fail (after side effects)", step)
}

// CheckCompensator returns an injected error if the compensator for step should fail
// on this attempt. Thread-safe; tracks attempt count internally.
func (c *SagaFaultConfig) CheckCompensator(step string) error {
	if c == nil || c.CompFailStep != step || c.CompFailTimes <= 0 {
		return nil
	}
	c.mu.Lock()
	n := c.attempts[step]
	c.attempts[step]++
	c.mu.Unlock()
	if n < c.CompFailTimes {
		return fmt.Errorf("fault injection: compensator %s forced fail (attempt %d/%d)", step, n+1, c.CompFailTimes)
	}
	return nil
}

// ApplyDelay sleeps if a delay is configured for step.
func (c *SagaFaultConfig) ApplyDelay(step string) {
	if c == nil || c.DelayStep != step || c.DelayMS <= 0 {
		return
	}
	time.Sleep(time.Duration(c.DelayMS) * time.Millisecond)
}
