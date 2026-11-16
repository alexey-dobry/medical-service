package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - New() - failed to parse conn url: %w", err)
	}

	pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres - New() - failed to create pool: %w", err)
	}

	// pgxpool подключается лениво, поэтому пингуем: иначе БД, ещё стартующая, даст ошибку только на первом запросе.
	for attempt := 1; attempt <= pg.connAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), pg.connTimeout)
		err = pg.Pool.Ping(ctx)
		cancel()
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts-attempt)

		time.Sleep(pg.connTimeout)
	}

	if err != nil {
		pg.Pool.Close()
		return nil, fmt.Errorf("postgres - New() - all attempts are timed out: %w", err)
	}

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
