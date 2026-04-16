package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventType is a domain-scoped event identifier.
// Format convention: "<domain>.<entity>.<action>"
// Examples: "patient.record.created", "lab.result.received"
type EventType string

// Message is the universal envelope for every event on the platform.
// It is intentionally domain-agnostic — it carries metadata + raw payload bytes.
type Message struct {
	ID              string          `json:"id"`
	Type            EventType       `json:"type"`
	Version         string          `json:"version"`
	Timestamp       time.Time       `json:"timestamp"`
	CorrelationID   string          `json:"correlation_id,omitempty"`
	CausationID     string          `json:"causation_id,omitempty"`
	ProducerService string          `json:"producer_service"`
	Payload         json.RawMessage `json:"payload"`
}

// NewMessage constructs a fully populated Message envelope.
// producerService: name of the service emitting the event (e.g. "patient-service")
// causationID: ID of the upstream message that triggered this one (empty if none)
func NewMessage(
	eventType EventType,
	version string,
	payload any,
	correlationID string,
	causationID string,
	producerService string,
) (*Message, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to marshal payload for event %q: %w", eventType, err)
	}

	return &Message{
		ID:              uuid.NewString(),
		Type:            eventType,
		Version:         version,
		Timestamp:       time.Now().UTC(),
		CorrelationID:   correlationID,
		CausationID:     causationID,
		ProducerService: producerService,
		Payload:         raw,
	}, nil
}

// Encode serializes the full envelope to JSON bytes for transport.
func (m *Message) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to encode message %q: %w", m.ID, err)
	}
	return b, nil
}

// DecodePayload deserializes the raw Payload field into the target struct.
// Usage: var p PatientCreatedPayload; msg.DecodePayload(&p)
func (m *Message) DecodePayload(target any) error {
	if err := json.Unmarshal(m.Payload, target); err != nil {
		return fmt.Errorf("kafka: failed to decode payload of event %q (type %q): %w",
			m.ID, m.Type, err)
	}
	return nil
}
