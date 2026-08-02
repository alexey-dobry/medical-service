package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Save(ctx context.Context, db DBTX, e Event) error {
	headers, err := json.Marshal(e.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}

	_, err = db.Exec(ctx,
		`INSERT INTO outbox_events (aggregate_id, event_type, payload, headers)
		 VALUES ($1, $2, $3, $4)`,
		e.AggregateID, e.EventType, e.Payload, headers,
	)

	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	return nil
}

func (s *Store) ProcessBatch(
	ctx context.Context,
	limit int,
	produce func(ctx context.Context, e Event) error,
) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT id, aggregate_id, event_type, payload, headers, created_at
		   FROM outbox_events
		  WHERE published_at IS NULL
		  ORDER BY id
		  LIMIT $1
		  FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return 0, fmt.Errorf("select unpublished: %w", err)
	}

	var events []Event
	for rows.Next() {
		var e Event
		var headers []byte
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.EventType,
			&e.Payload, &headers, &e.CreatedAt); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan: %w", err)
		}
		if len(headers) > 0 {
			_ = json.Unmarshal(headers, &e.Headers)
		}
		events = append(events, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate rows: %w", err)
	}
	if len(events) == 0 {
		return 0, nil
	}

	ids := make([]int64, 0, len(events))
	for _, e := range events {
		if err := produce(ctx, e); err != nil {
			return 0, fmt.Errorf("produce event %d: %w", e.ID, err)
		}
		ids = append(ids, e.ID)
	}

	if _, err = tx.Exec(ctx,
		`UPDATE outbox_events SET published_at = now() WHERE id = ANY($1)`,
		ids,
	); err != nil {
		return 0, fmt.Errorf("mark published: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(events), nil
}
