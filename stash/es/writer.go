package es

import (
	"context"

	"github.com/olivere/elastic/v7"
	"github.com/tal-tech/go-stash/stash/config"
	"github.com/tal-tech/go-zero/core/executors"
	"github.com/tal-tech/go-zero/core/logx"
)

type (
	Writer struct {
		docType  string
		client   *elastic.Client
		inserter *executors.ChunkExecutor
	}

	valueWithIndex struct {
		index string
		val   string
	}
)

func NewWriter(c config.ElasticSearchConf) (*Writer, error) {
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(c.Hosts...),
		elastic.SetGzip(c.Compress),
	)
	if err != nil {
		return nil, err
	}

	writer := Writer{
		docType: c.DocType,
		client:  client,
	}
	writer.inserter = executors.NewChunkExecutor(writer.execute, executors.WithChunkBytes(c.MaxChunkBytes))
	return &writer, nil
}

func (w *Writer) Write(index, val string) error {
	return w.inserter.Add(valueWithIndex{
		index: index,
		val:   val,
	}, len(val))
}

func (w *Writer) execute(vals []interface{}) {
	var bulk = w.client.Bulk()
	for _, val := range vals {
		pair := val.(valueWithIndex)
		req := elastic.NewBulkIndexRequest().Index(pair.index).Type(w.docType).Doc(pair.val)
		bulk.Add(req)
	}
	resp, err := bulk.Do(context.Background())
	if err != nil {
		logx.Error(err)
		return
	}

	// bulk error in docs will report in response items
	if !resp.Errors {
		return
	}

	for _, imap := range resp.Items {
		for _, item := range imap {
			if item.Error == nil {
				continue
			}

			logx.Error(item.Error)
		}
	}
}
