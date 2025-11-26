package filter

import (
	"time"
)

const (
	timestampFormat = "2006-01-02T15:04:05.000Z"
	timestampKey    = "@timestamp"
)

// TimestampFilter adds a @timestamp field to the message if it doesn't exist
func TimestampFilter() FilterFunc {
	return func(m map[string]interface{}) map[string]interface{} {
		if _, ok := m[timestampKey]; !ok {
			m[timestampKey] = time.Now().UTC().Format(timestampFormat)
		}
		return m
	}
}
