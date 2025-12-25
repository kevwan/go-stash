package config

import (
	"context"
	"regexp"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// FetchMatchingTopics fetches topics from Kafka that match the given pattern.
// If pattern is empty, it returns the Topics list from the config.
func FetchMatchingTopics(c KafkaConf) ([]string, error) {
	if c.TopicsPattern == "" {
		return c.Topics, nil
	}

	// Compile the regex pattern
	pattern, err := regexp.Compile(c.TopicsPattern)
	if err != nil {
		return nil, err
	}

	// Try connecting to each broker until one succeeds
	var conn *kafka.Conn
	var lastErr error
	for _, broker := range c.Brokers {
		conn, err = kafka.DialContext(context.Background(), "tcp", broker)
		if err == nil {
			break
		}
		lastErr = err
	}
	if conn == nil {
		return nil, lastErr
	}
	defer conn.Close()

	// Set deadline for the operation
	deadline := 10 * time.Second
	if err := conn.SetDeadline(time.Now().Add(deadline)); err != nil {
		return nil, err
	}

	// Fetch partition metadata (includes all topics)
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, err
	}

	// Extract unique topics that match the pattern
	topicSet := make(map[string]struct{})
	for _, partition := range partitions {
		if pattern.MatchString(partition.Topic) {
			topicSet[partition.Topic] = struct{}{}
		}
	}

	// Convert to slice
	var topics []string
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	logx.Infof("Matched %d topics with pattern '%s': %v", len(topics), c.TopicsPattern, topics)
	return topics, nil
}
