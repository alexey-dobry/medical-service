package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// HandlerFunc processes a decoded Message envelope.
// Returning a non-nil error prevents offset commit → message will be redelivered.
type HandlerFunc func(ctx context.Context, msg *Message) error

// RawHandlerFunc processes every message regardless of EventType.
// Used by services like audit that consume all events universally.
type RawHandlerFunc func(ctx context.Context, msg *Message) error

// Consumer wraps kafka-go's Reader with per-EventType handler dispatch.
type Consumer struct {
	reader   *kafkago.Reader
	handlers map[EventType]HandlerFunc
	logger   logger.Logger
	topic    string
}

// NewConsumer constructs a Consumer subscribed to a single topic.
func NewConsumer(cfg Config, topic string, logger logger.Logger) (*Consumer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       cfg.Consumer.MinBytes,
		MaxBytes:       cfg.Consumer.MaxBytes,
		MaxWait:        cfg.Consumer.MaxWait,
		CommitInterval: cfg.Consumer.CommitInterval,
		StartOffset:    cfg.Consumer.StartOffset,
		Logger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debugf(msg, args...)
		}),
		ErrorLogger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Errorf(msg, args...)
		}),
	})

	return &Consumer{
		reader:   r,
		handlers: make(map[EventType]HandlerFunc),
		logger:   logger,
		topic:    topic,
	}, nil
}

// RegisterHandler maps an EventType to a handler function.
// Must be called before Start(). Safe to call multiple times for different event types.
func (c *Consumer) RegisterHandler(eventType EventType, fn HandlerFunc) {
	c.handlers[eventType] = fn
	c.logger.Info("kafka: handler registered",
		zap.String("topic", c.topic),
		zap.String("event_type", string(eventType)),
	)
}

// Start begins the consume-dispatch loop. Blocks until ctx is cancelled.
//
// Per-message lifecycle:
//  1. Fetch raw message from broker (no offset commit yet)
//  2. Decode JSON envelope into *Message
//  3. Look up registered handler by EventType
//  4. Call handler — if it returns error, skip commit (message redelivered)
//  5. Commit offset only after successful handler return
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("kafka: consumer starting",
		zap.String("topic", c.topic),
		zap.Int("handlers", len(c.handlers)),
	)

	for {
		rawMsg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("kafka: consumer shutting down (context cancelled)",
					zap.String("topic", c.topic),
				)
				return nil
			}
			c.logger.Error("kafka: fetch error", zap.Error(err), zap.String("topic", c.topic))
			continue
		}

		if err := c.dispatch(ctx, rawMsg); err != nil {
			// dispatch logs internally; we don't commit so the message is retried
			continue
		}

		if err := c.reader.CommitMessages(ctx, rawMsg); err != nil {
			c.logger.Error("kafka: failed to commit offset",
				zap.Int64("offset", rawMsg.Offset),
				zap.Error(err),
			)
		}
	}
}

// StartRaw runs a consume loop that passes every message to a single universal handler,
// regardless of EventType. Designed for audit/logging consumers.
// Malformed envelopes are still committed to avoid poison-pill stalls.
func (c *Consumer) StartRaw(ctx context.Context, handler RawHandlerFunc) error {
	c.logger.Info("kafka: raw consumer starting", zap.String("topic", c.topic))

	for {
		rawMsg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("kafka: fetch error", zap.Error(err))
			continue
		}

		var msg Message
		if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
			c.logger.Error("kafka: malformed envelope, committing and skipping",
				zap.Error(err),
				zap.Int64("offset", rawMsg.Offset),
			)
			_ = c.reader.CommitMessages(ctx, rawMsg)
			continue
		}

		if err := handler(ctx, &msg); err != nil {
			c.logger.Error("kafka: raw handler failed, will retry",
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			continue // don't commit
		}

		_ = c.reader.CommitMessages(ctx, rawMsg)
	}
}

// Lag returns the number of messages this consumer is behind the latest offset.
// Expose via /metrics or /health for monitoring.
func (c *Consumer) Lag() int64 {
	return c.reader.Lag()
}

// Close gracefully shuts down the reader. Always defer this.
func (c *Consumer) Close() error {
	c.logger.Info("kafka: consumer closing", zap.String("topic", c.topic))
	return c.reader.Close()
}

// dispatch decodes a raw Kafka message and routes it to the correct handler.
func (c *Consumer) dispatch(ctx context.Context, rawMsg kafkago.Message) error {
	var msg Message
	if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
		c.logger.Error("kafka: failed to decode envelope — committing to skip poison pill",
			zap.Error(err),
			zap.Int64("offset", rawMsg.Offset),
			zap.String("topic", c.topic),
		)
		// Commit malformed message to prevent infinite retry loop
		_ = c.reader.CommitMessages(ctx, rawMsg)
		return nil // not a retryable error
	}

	c.logger.Info("kafka: message received",
		zap.String("event_type", string(msg.Type)),
		zap.String("message_id", msg.ID),
		zap.String("correlation_id", msg.CorrelationID),
		zap.Int64("offset", rawMsg.Offset),
	)

	handler, ok := c.handlers[msg.Type]
	if !ok {
		c.logger.Debug("kafka: no handler registered — skipping",
			zap.String("event_type", string(msg.Type)),
		)
		// Intentionally commit — this service doesn't care about this event type
		_ = c.reader.CommitMessages(ctx, rawMsg)
		return nil
	}

	if err := handler(ctx, &msg); err != nil {
		c.logger.Error("kafka: handler returned error — will redeliver",
			zap.String("event_type", string(msg.Type)),
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
		return &HandlerError{EventType: msg.Type, MessageID: msg.ID, Err: err}
	}

	return nil
}
