package handler

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/tal-tech/go-stash/stash/es"
	"github.com/tal-tech/go-stash/stash/filter"
)

type MessageHandler struct {
	writer  *es.Writer
	filters []filter.FilterFunc
}

func NewHandler(writer *es.Writer) *MessageHandler {
	return &MessageHandler{
		writer: writer,
	}
}

func (mh *MessageHandler) AddFilters(filters ...filter.FilterFunc) {
	for _, f := range filters {
		mh.filters = append(mh.filters, f)
	}
}

func (mh *MessageHandler) Consume(_, val string) error {
	m := make(map[string]interface{})
	if err := jsoniter.Unmarshal([]byte(val), &m); err != nil {
		return err
	}

	for _, proc := range mh.filters {
		if m = proc(m); m == nil {
			return nil
		}
	}

	return mh.writer.Write(m)
}
