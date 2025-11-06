package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/service"
)

func TestFetchMatchingTopics_WithExplicitTopics(t *testing.T) {
	conf := KafkaConf{
		Topics: []string{"topic1", "topic2", "topic3"},
	}

	topics, err := FetchMatchingTopics(conf)
	assert.NoError(t, err)
	assert.Equal(t, []string{"topic1", "topic2", "topic3"}, topics)
}

func TestFetchMatchingTopics_EmptyPattern(t *testing.T) {
	conf := KafkaConf{
		Topics:        []string{"topic1", "topic2"},
		TopicsPattern: "",
	}

	topics, err := FetchMatchingTopics(conf)
	assert.NoError(t, err)
	assert.Equal(t, []string{"topic1", "topic2"}, topics)
}

func TestFetchMatchingTopics_InvalidPattern(t *testing.T) {
	conf := KafkaConf{
		ServiceConf:   service.ServiceConf{},
		Brokers:       []string{"localhost:9092"},
		TopicsPattern: "[invalid-regex",
	}

	_, err := FetchMatchingTopics(conf)
	assert.Error(t, err)
}
