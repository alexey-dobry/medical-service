// pkg/outbox/relay.go
package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
)

type Relay struct {
	store    *Store
	producer Producer
	interval time.Duration
	batch    int
	log      logger.Logger
}

type Option func(*Relay)

func WithInterval(d time.Duration) Option { return func(r *Relay) { r.interval = d } }
func WithBatchSize(n int) Option          { return func(r *Relay) { r.batch = n } }

func NewRelay(store *Store, producer Producer, log logger.Logger, opts ...Option) *Relay {
	r := &Relay{
		store:    store,
		producer: producer,
		interval: 200 * time.Millisecond,
		batch:    100,
		log:      log,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for {
				n, err := r.store.ProcessBatch(ctx, r.batch, r.producer.Publish)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					r.log.Error("outbox relay batch failed", "err", err)
					break
				}
				if n < r.batch {
					break
				}
			}
		}
	}
}
