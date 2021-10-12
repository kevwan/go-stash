package handler

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kevwan/go-stash/stash/es"
	"github.com/kevwan/go-stash/stash/filter"
)

type MessageHandler struct {
	writer  *es.Writer
	indexer *es.Index
	filters []filter.FilterFunc
}

func NewHandler(writer *es.Writer, indexer *es.Index) *MessageHandler {
	return &MessageHandler{
		writer:  writer,
		indexer: indexer,
	}
}

func (mh *MessageHandler) AddFilters(filters ...filter.FilterFunc) {
	for _, f := range filters {
		mh.filters = append(mh.filters, f)
	}
}

func (mh *MessageHandler) Consume(_, val string) error {
	var m map[string]interface{}
	if err := jsoniter.Unmarshal([]byte(val), &m); err != nil {
		return err
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
