package kafka

import (
	"errors"
	"fmt"
)

var (
	ErrProducerClosed      = errors.New("kafka: producer is closed")
	ErrConsumerClosed      = errors.New("kafka: consumer is closed")
	ErrSerializationFailed = errors.New("kafka: message serialization failed")
	ErrInvalidConfig       = errors.New("kafka: invalid configuration")
)

// PublishError carries the context of a failed publish operation.
type PublishError struct {
	Topic     string
	EventType EventType
	MessageID string
	Err       error
}

func (e *PublishError) Error() string {
	return fmt.Sprintf(
		"kafka: failed to publish message %q (event=%q, topic=%q): %v",
		e.MessageID, e.EventType, e.Topic, e.Err,
	)
}

func (e *PublishError) Unwrap() error { return e.Err }

// HandlerError carries the context of a failed handler execution.
type HandlerError struct {
	EventType EventType
	MessageID string
	Err       error
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf(
		"kafka: handler failed for message %q (event=%q): %v",
		e.MessageID, e.EventType, e.Err,
	)
}

func (e *HandlerError) Unwrap() error { return e.Err }
