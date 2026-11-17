// Package elasticsearch wraps the official client with the same conventions
// as pkg/postgres and pkg/redis: functional options, connection retries on
// start, Ping for readiness probes and Close.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const (
	_defaultMaxRetries      = 3
	_defaultResponseTimeout = 10 * time.Second
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
)

var ErrNotFound = errors.New("elasticsearch: not found")

type Elasticsearch struct {
	username        string
	password        string
	maxRetries      int
	responseTimeout time.Duration
	connAttempts    int
	connTimeout     time.Duration

	Client *elasticsearch.Client
}

func New(addresses []string, opts ...Option) (*Elasticsearch, error) {
	es := &Elasticsearch{
		maxRetries:      _defaultMaxRetries,
		responseTimeout: _defaultResponseTimeout,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(es)
	}

	if len(addresses) == 0 {
		return nil, errors.New("elasticsearch - New() - no addresses given")
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:  addresses,
		Username:   es.username,
		Password:   es.password,
		MaxRetries: es.maxRetries,
		Transport: &http.Transport{
			ResponseHeaderTimeout: es.responseTimeout,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch - New() - failed to create client: %w", err)
	}
	es.Client = client

	for attempt := 1; attempt <= es.connAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), es.connTimeout)
		err = es.Ping(ctx)
		cancel()
		if err == nil {
			break
		}

		log.Printf("Elasticsearch is trying to connect, attempts left: %d", es.connAttempts-attempt)

		time.Sleep(es.connTimeout)
	}

	if err != nil {
		return nil, fmt.Errorf("elasticsearch - New() - all attempts are timed out: %w", err)
	}

	return es, nil
}

// ping для /readyz
func (e *Elasticsearch) Ping(ctx context.Context) error {
	res, err := e.Client.Ping(e.Client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elasticsearch - Ping(): %w", err)
	}
	defer res.Body.Close()

	return CheckResponse(res)
}

func (e *Elasticsearch) EnsureIndex(ctx context.Context, index string, mapping any) error {
	exists, err := e.Client.Indices.Exists([]string{index}, e.Client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elasticsearch - EnsureIndex() - check %q: %w", index, err)
	}
	defer exists.Body.Close()

	switch exists.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
	default:
		return fmt.Errorf("elasticsearch - EnsureIndex() - check %q: %w", index, CheckResponse(exists))
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("elasticsearch - EnsureIndex() - marshal mapping: %w", err)
	}

	created, err := e.Client.Indices.Create(
		index,
		e.Client.Indices.Create.WithBody(bytes.NewReader(body)),
		e.Client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch - EnsureIndex() - create %q: %w", index, err)
	}
	defer created.Body.Close()

	if created.IsError() {
		cerr := CheckResponse(created)
		var esErr *ResponseError
		if errors.As(cerr, &esErr) && esErr.Type == "resource_already_exists_exception" {
			return nil
		}
		return fmt.Errorf("elasticsearch - EnsureIndex() - create %q: %w", index, cerr)
	}

	return nil
}

func (e *Elasticsearch) Close() error {
	if e.Client == nil {
		return nil
	}
	return e.Client.Close(context.Background())
}

type ResponseError struct {
	StatusCode int
	Type       string
	Reason     string
}

func (e *ResponseError) Error() string {
	if e.Type == "" {
		return fmt.Sprintf("elasticsearch: HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("elasticsearch: HTTP %d %s: %s", e.StatusCode, e.Type, e.Reason)
}

func (e *ResponseError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}

func CheckResponse(res *esapi.Response) error {
	if !res.IsError() {
		return nil
	}

	respErr := &ResponseError{StatusCode: res.StatusCode}

	var body struct {
		Error json.RawMessage `json:"error"`
	}
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if json.Unmarshal(raw, &body) == nil && len(body.Error) > 0 {
		var detailed struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		}
		if json.Unmarshal(body.Error, &detailed) == nil {
			respErr.Type, respErr.Reason = detailed.Type, detailed.Reason
		} else {
			// Some endpoints return "error" as a plain string.
			var s string
			if json.Unmarshal(body.Error, &s) == nil {
				respErr.Reason = s
			}
		}
	}

	return respErr
}
