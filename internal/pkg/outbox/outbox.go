// pkg/outbox/outbox.go
package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Event struct {
	ID          int64
	AggregateID string            // ключ партиции (user_id)
	EventType   string            // "UserCreationRequested"
	Payload     []byte            // proto.Marshal(...)
	Headers     map[string]string // saga_id, traceparent
	CreatedAt   time.Time
	PublishedAt *time.Time
}

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Producer interface {
	Publish(ctx context.Context, e Event) error
}
