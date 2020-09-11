package es

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/olivere/elastic"
	"github.com/tal-tech/go-stash/stash/config"
	"github.com/tal-tech/go-zero/core/executors"
	"github.com/tal-tech/go-zero/core/logx"
)

type (
	Writer struct {
		docType  string
		client   *elastic.Client
		indexer  *Index
		inserter *executors.ChunkExecutor
	}

	valueWithIndex struct {
		index string
		val   string
	}
)

func NewWriter(c config.ElasticSearchConf, indexer *Index) (*Writer, error) {
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
		indexer: indexer,
	}
	writer.inserter = executors.NewChunkExecutor(writer.execute, executors.WithChunkBytes(c.MaxChunkBytes))
	return &writer, nil
}

func (w *Writer) Write(m map[string]interface{}) error {
	bs, err := jsoniter.Marshal(m)
	if err != nil {
		return err
	}

	index := w.indexer.GetIndex(m)
	val := string(bs)
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
	_, err := bulk.Do(context.Background())
	if err != nil {
		logx.Error(err)
	}
}
