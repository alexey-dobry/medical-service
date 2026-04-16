package kafka

import (
	"fmt"
	"time"
)

type Config struct {
	Brokers  []string
	GroupID  string
	ClientID string
	Producer ProducerConfig
	Consumer ConsumerConfig
}

type ProducerConfig struct {
	BatchTimeout time.Duration
	MaxAttempts  int
	RequiredAcks RequiredAcks
	Async        bool
}

type ConsumerConfig struct {
	StartOffset    int64
	MinBytes       int
	MaxBytes       int
	MaxWait        time.Duration
	CommitInterval time.Duration
}

type RequiredAcks int

const (
	RequireNone RequiredAcks = 0
	RequireOne  RequiredAcks = 1
	RequireAll  RequiredAcks = -1
)

func DefaultConfig(brokers []string, groupID, clientID string) Config {
	return Config{
		Brokers:  brokers,
		GroupID:  groupID,
		ClientID: clientID,
		Producer: ProducerConfig{
			BatchTimeout: 10 * time.Millisecond,
			MaxAttempts:  5,
			RequiredAcks: RequireAll,
			Async:        false,
		},
		Consumer: ConsumerConfig{
			StartOffset:    -2, // FirstOffset
			MinBytes:       1,
			MaxBytes:       10 << 20, // 10MB
			MaxWait:        500 * time.Millisecond,
			CommitInterval: time.Second,
		},
	}
}

func (c Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("at least one broker is required")
	}
	if c.GroupID == "" {
		return fmt.Errorf("consumer group ID is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	return nil
}
