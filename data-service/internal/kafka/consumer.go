package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"strings"
)

const (
	sessionTimeOut = 7000 // ms
	noTimeout      = -1
)

type Consumer struct {
	consumer       *kafka.Consumer
	stop           bool
	consumerNumber int
}

func NewConsumer(address []string, topic, consumerGroup string) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"group.id":           consumerGroup,
		"session.timeout.ms": sessionTimeOut,
		"auto.offset.reset":  "earliest",
	}
	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error with new consumer: %w", err)
	}
	if err = c.Subscribe(topic, nil); err != nil {
		return nil, err
	}
	return &Consumer{
		consumer: c,
		stop:     false,
	}, nil
}

func (c *Consumer) StartWithFunc(hf func(msg []byte, offset kafka.Offset) error) {
	for !c.stop {
		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Fatal(err)
		}
		if kafkaMsg == nil {
			continue
		}
		if err := hf(kafkaMsg.Value, kafkaMsg.TopicPartition.Offset); err != nil {
			log.Printf("handler error: %v", err)
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	log.Print("Commited offset")
	return c.consumer.Close()
}
