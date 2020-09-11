package es

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic"
	"github.com/tal-tech/go-zero/core/fx"
	"github.com/tal-tech/go-zero/core/lang"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/syncx"
)

const (
	sharedCallsKey  = "ensureIndex"
	timestampFormat = "2006-01-02T15:04:05.000Z"
	timestampKey    = "@timestamp"
)

const (
	stateNormal = iota
	stateWrap
	stateDot
)

type (
	IndexFormat func(m map[string]interface{}) string
	IndexFunc   func() string

	Index struct {
		client      *elastic.Client
		indexFormat IndexFormat
		indices     map[string]lang.PlaceholderType
		lock        sync.RWMutex
		sharedCalls syncx.SharedCalls
	}
)

func NewIndex(client *elastic.Client, indexFormat string, loc *time.Location) *Index {
	var formatter func(map[string]interface{}) string
	format, attrs := getFormat(indexFormat)
	if len(attrs) > 0 {
		formatter = func(m map[string]interface{}) string {
			var vals []interface{}
			for _, attr := range attrs {
				if val, ok := m[attr]; ok {
					vals = append(vals, val)
				}
			}
			return getTime(m).In(loc).Format(fmt.Sprintf(format, vals...))
		}
	} else {
		formatter = func(m map[string]interface{}) string {
			return getTime(m).In(loc).Format(format)
		}
	}

	return &Index{
		client:      client,
		indexFormat: formatter,
		indices:     make(map[string]lang.PlaceholderType),
		sharedCalls: syncx.NewSharedCalls(),
	}
}

func (idx *Index) GetIndex(m map[string]interface{}) string {
	index := idx.indexFormat(m)
	idx.lock.RLock()
	if _, ok := idx.indices[index]; ok {
		idx.lock.RUnlock()
		return index
	}

	idx.lock.RUnlock()
	if err := idx.ensureIndex(index); err != nil {
		logx.Error(err)
	}
	return index
}

func (idx *Index) ensureIndex(index string) error {
	_, err := idx.sharedCalls.Do(sharedCallsKey, func() (i interface{}, err error) {
		idx.lock.Lock()
		defer idx.lock.Unlock()

		if _, ok := idx.indices[index]; ok {
			return nil, nil
		}

		existsService := elastic.NewIndicesExistsService(idx.client)
		existsService.Index([]string{index})
		exist, err := existsService.Do(context.Background())
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, nil
		}

		createService := idx.client.CreateIndex(index)
		if err := fx.DoWithRetries(func() error {
			// is it necessary to check the result?
			_, err := createService.Do(context.Background())
			return err
		}); err != nil {
			return nil, err
		}

		idx.indices[index] = lang.Placeholder
		return nil, nil
	})
	return err
}

func getTime(m map[string]interface{}) time.Time {
	if ti, ok := m[timestampKey]; ok {
		if ts, ok := ti.(string); ok {
			if t, err := time.Parse(timestampFormat, ts); err == nil {
				return t
			}
		}
	}

	return time.Now()
}

func getFormat(indexFormat string) (format string, attrs []string) {
	var state = stateNormal
	var builder strings.Builder
	var keyBuf strings.Builder
	for _, ch := range indexFormat {
		switch ch {
		case '{':
			state = stateWrap
		case '.':
			if state == stateWrap {
				state = stateDot
			} else {
				builder.WriteRune(ch)
			}
		case '}':
			state = stateNormal
			if keyBuf.Len() > 0 {
				attrs = append(attrs, keyBuf.String())
				builder.WriteString("%s")
			}
		default:
			if state == stateDot {
				keyBuf.WriteRune(ch)
			} else {
				builder.WriteRune(ch)
			}
		}
	}

	format = builder.String()
	return
}
