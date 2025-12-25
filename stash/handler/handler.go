package handler

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/kevwan/go-stash/stash/es"
	"github.com/kevwan/go-stash/stash/filter"
)

type MessageHandler struct {
	writer  *es.Writer
	indexer *es.Index
	filters []filter.FilterFunc
	topic   string
}

func NewHandler(writer *es.Writer, indexer *es.Index) *MessageHandler {
	return &MessageHandler{
		writer:  writer,
		indexer: indexer,
	}
}

func NewHandlerWithTopic(writer *es.Writer, indexer *es.Index, topic string) *MessageHandler {
	return &MessageHandler{
		writer:  writer,
		indexer: indexer,
		topic:   topic,
	}
}

func (mh *MessageHandler) AddFilters(filters ...filter.FilterFunc) {
	for _, f := range filters {
		mh.filters = append(mh.filters, f)
	}
}

// ensureMetadataStructure ensures the @metadata.kafka structure exists in the message.
func ensureMetadataStructure(m map[string]interface{}) map[string]interface{} {
	if _, exists := m["@metadata"]; !exists {
		m["@metadata"] = make(map[string]interface{})
	}
	metadata, ok := m["@metadata"].(map[string]interface{})
	if !ok {
		return nil
	}
	if _, exists := metadata["kafka"]; !exists {
		metadata["kafka"] = make(map[string]interface{})
	}
	kafkaMeta, ok := metadata["kafka"].(map[string]interface{})
	if !ok {
		return nil
	}
	return kafkaMeta
}

func (mh *MessageHandler) Consume(_ context.Context, _, val string) error {
	var m map[string]interface{}
	if err := jsoniter.Unmarshal([]byte(val), &m); err != nil {
		return err
	}

	// Inject Kafka metadata if topic is set
	if mh.topic != "" {
		if kafkaMeta := ensureMetadataStructure(m); kafkaMeta != nil {
			kafkaMeta["topic"] = mh.topic
		}
	}

	index := mh.indexer.GetIndex(m)
	for _, proc := range mh.filters {
		if m = proc(m); m == nil {
			return nil
		}
	}

	bs, err := jsoniter.Marshal(m)
	if err != nil {
		return err
	}

	return mh.writer.Write(index, string(bs))
}

