package kafka

import (
	"context"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// BatchMessage pairs a typed message with its partition routing key.
type BatchMessage struct {
	Msg *Message
	Key string
}

// Producer is a generic, domain-agnostic Kafka writer.
// It knows nothing about topics, event types, or payload schemas —
// those concerns belong to each microservice.
type Producer struct {
	writer *kafkago.Writer
	logger logger.Logger
	cfg    ProducerConfig
}

// NewProducer constructs and validates a Producer from config.
func NewProducer(cfg Config, logger logger.Logger) (*Producer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(cfg.Brokers...),
		Balancer:               &kafkago.Hash{}, // key-based partitioning for ordering
		MaxAttempts:            cfg.Producer.MaxAttempts,
		BatchTimeout:           cfg.Producer.BatchTimeout,
		RequiredAcks:           kafkago.RequiredAcks(cfg.Producer.RequiredAcks),
		Async:                  cfg.Producer.Async,
		AllowAutoTopicCreation: false, // topics must be explicitly created in production
		Logger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debugf(msg, args...)
		}),
		ErrorLogger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Errorf(msg, args...)
		}),
	}

	return &Producer{writer: w, logger: logger, cfg: cfg.Producer}, nil
}

// Publish serializes and writes a single message to the given topic.
// key is the partition routing key (e.g. patientID, appointmentID).
// Use the same key for all events belonging to the same entity to guarantee ordering.
func (p *Producer) Publish(ctx context.Context, topic string, msg *Message, key string) error {
	encoded, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerializationFailed, err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: encoded,
		Headers: []kafkago.Header{
			{Key: "event-type", Value: []byte(msg.Type)},
			{Key: "version", Value: []byte(msg.Version)},
			{Key: "correlation-id", Value: []byte(msg.CorrelationID)},
			{Key: "causation-id", Value: []byte(msg.CausationID)},
			{Key: "producer", Value: []byte(msg.ProducerService)},
		},
	})
	if err != nil {
		return &PublishError{
			Topic:     topic,
			EventType: msg.Type,
			MessageID: msg.ID,
			Err:       err,
		}
	}

	p.logger.Info("kafka: published",
		zap.String("topic", topic),
		zap.String("event_type", string(msg.Type)),
		zap.String("message_id", msg.ID),
		zap.String("key", key),
		zap.String("correlation_id", msg.CorrelationID),
	)

	return nil
}

// PublishBatch writes multiple messages to the same topic in one broker round-trip.
// More efficient than calling Publish in a loop.
func (p *Producer) PublishBatch(ctx context.Context, topic string, batch []BatchMessage) error {
	msgs := make([]kafkago.Message, 0, len(batch))

	for _, bm := range batch {
		encoded, err := bm.Msg.Encode()
		if err != nil {
			return fmt.Errorf("%w: message_id=%s: %v", ErrSerializationFailed, bm.Msg.ID, err)
		}
		msgs = append(msgs, kafkago.Message{
			Topic: topic,
			Key:   []byte(bm.Key),
			Value: encoded,
		})
	}

	if err := p.writer.WriteMessages(ctx, msgs...); err != nil {
		return &PublishError{Topic: topic, Err: err}
	}

	p.logger.Info("kafka: batch published",
		zap.String("topic", topic),
		zap.Int("count", len(batch)),
	)

	return nil
}

// Close flushes all pending messages and releases the writer.
// Always defer this after creating a Producer.
func (p *Producer) Close() error {
	p.logger.Info("kafka: producer closing")
	return p.writer.Close()
}
