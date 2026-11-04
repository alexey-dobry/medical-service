package service

import (
	"context"
	"fmt"
	"time"
)

// RetryPolicy описывает экспоненциальный backoff: InitialDelay, 2*InitialDelay,
// 4*InitialDelay ... с потолком MaxDelay и не более Attempts вызовов.
type RetryPolicy struct {
	Attempts     int           `yaml:"attempts" env-default:"5"`
	InitialDelay time.Duration `yaml:"initial_delay" env-default:"200ms"`
	MaxDelay     time.Duration `yaml:"max_delay" env-default:"5s"`
	// Timeout ограничивает всю компенсацию, включая все попытки.
	Timeout time.Duration `yaml:"timeout" env-default:"30s"`
}

// retry вызывает fn, пока она не завершится успешно, не исчерпаются попытки
// или не отменится ctx. Возвращает последнюю ошибку.
func retry(ctx context.Context, policy RetryPolicy, fn func(context.Context) error) error {
	attempts := max(policy.Attempts, 1)
	delay := policy.InitialDelay

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = fn(ctx); err == nil {
			return nil
		}

		if attempt == attempts {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("%w (last error: %w)", ctx.Err(), err)
		case <-timer.C:
		}

		delay = min(delay*2, policy.MaxDelay)
	}

	return fmt.Errorf("all %d attempts failed: %w", attempts, err)
}
