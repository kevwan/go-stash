package handler

import (
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kevwan/go-stash/stash/filter"
	"github.com/stretchr/testify/assert"
)

func TestTimestampFilterIntegration(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectField  bool
		expectValue  string
	}{
		{
			name:        "message without timestamp should get one",
			input:       `{"message":"test","level":"info"}`,
			expectField: true,
		},
		{
			name:        "message with timestamp should keep it",
			input:       `{"message":"test","@timestamp":"2022-01-01T00:00:00.000Z"}`,
			expectField: true,
			expectValue: "2022-01-01T00:00:00.000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the timestamp filter like it would be in production
			timestampFilter := filter.TimestampFilter()

			var m map[string]interface{}
			err := jsoniter.Unmarshal([]byte(tt.input), &m)
			assert.NoError(t, err)

			// Apply filter
			result := timestampFilter(m)
			assert.NotNil(t, result)

			// Verify timestamp field exists
			if tt.expectField {
				timestamp, exists := result["@timestamp"]
				assert.True(t, exists, "@timestamp field should exist")
				assert.NotEmpty(t, timestamp)

				// If we expect a specific value, check it
				if tt.expectValue != "" {
					assert.Equal(t, tt.expectValue, timestamp)
				} else {
					// Verify it's a valid timestamp format (ISO 8601)
					const iso8601Format = "2006-01-02T15:04:05.000Z"
					timestampStr, ok := timestamp.(string)
					assert.True(t, ok)
					_, err := time.Parse(iso8601Format, timestampStr)
					assert.NoError(t, err, "Timestamp should be in ISO 8601 format")
				}
			}
		})
	}
}
