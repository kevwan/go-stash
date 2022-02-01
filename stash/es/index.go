package es

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/vjeantet/jodaTime"
	"github.com/zeromicro/go-zero/core/fx"
	"github.com/zeromicro/go-zero/core/lang"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
)

const (
	timestampFormat = "2006-01-02T15:04:05.000Z"
	timestampKey    = "@timestamp"
	leftBrace       = '{'
	rightBrace      = '}'
	dot             = '.'
)

const (
	stateNormal = iota
	stateWrap
	stateVar
	stateDot
)

type (
	IndexFormat func(m map[string]interface{}) string
	IndexFunc   func() string

	Index struct {
		client       *elastic.Client
		indexFormat  IndexFormat
		indices      map[string]lang.PlaceholderType
		lock         sync.RWMutex
		singleFlight syncx.SingleFlight
	}
)

func NewIndex(client *elastic.Client, indexFormat string, loc *time.Location) *Index {
	return &Index{
		client:       client,
		indexFormat:  buildIndexFormatter(indexFormat, loc),
		indices:      make(map[string]lang.PlaceholderType),
		singleFlight: syncx.NewSingleFlight(),
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
	_, err := idx.singleFlight.Do(index, func() (i interface{}, err error) {
		idx.lock.Lock()
		defer idx.lock.Unlock()

		if _, ok := idx.indices[index]; ok {
			return nil, nil
		}

		defer func() {
			if err == nil {
				idx.indices[index] = lang.Placeholder
			}
		}()

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
		if err := fx.DoWithRetry(func() error {
			// is it necessary to check the result?
			_, err := createService.Do(context.Background())
			return err
		}); err != nil {
			return nil, err
		}

		return nil, nil
	})
	return err
}

func buildIndexFormatter(indexFormat string, loc *time.Location) func(map[string]interface{}) string {
	format, attrs, timePos := getFormat(indexFormat)
	if len(attrs) == 0 {
		return func(m map[string]interface{}) string {
			return format
		}
	}

	return func(m map[string]interface{}) string {
		var vals []interface{}
		for i, attr := range attrs {
			if i == timePos {
				vals = append(vals, formatTime(attr, getTime(m).In(loc)))
				continue
			}

			if val, ok := m[attr]; ok {
				vals = append(vals, val)
			} else {
				vals = append(vals, "")
			}
		}
		return fmt.Sprintf(format, vals...)
	}
}

func formatTime(format string, t time.Time) string {
	return jodaTime.Format(format, t)
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

func getFormat(indexFormat string) (format string, attrs []string, timePos int) {
	var state = stateNormal
	var builder strings.Builder
	var keyBuf strings.Builder
	timePos = -1
	writeHolder := func() {
		if keyBuf.Len() > 0 {
			attrs = append(attrs, keyBuf.String())
			keyBuf.Reset()
			builder.WriteString("%s")
		}
	}

	for _, ch := range indexFormat {
		switch state {
		case stateNormal:
			switch ch {
			case leftBrace:
				state = stateWrap
			default:
				builder.WriteRune(ch)
			}
		case stateWrap:
			switch ch {
			case leftBrace:
				state = stateVar
			case dot:
				state = stateDot
				keyBuf.Reset()
			case rightBrace:
				state = stateNormal
				timePos = len(attrs)
				writeHolder()
			default:
				keyBuf.WriteRune(ch)
			}
		case stateVar:
			switch ch {
			case rightBrace:
				state = stateWrap
			default:
				keyBuf.WriteRune(ch)
			}
		case stateDot:
			switch ch {
			case rightBrace:
				state = stateNormal
				writeHolder()
			default:
				keyBuf.WriteRune(ch)
			}
		default:
			builder.WriteRune(ch)
		}
	}

	format = builder.String()
	return
}
