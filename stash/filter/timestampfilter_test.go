package filter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		hasField bool
	}{
		{
			name:     "add timestamp when not present",
			input:    map[string]interface{}{"field1": "value1"},
			hasField: false,
		},
		{
			name:     "preserve existing timestamp",
			input:    map[string]interface{}{"@timestamp": "2022-01-01T00:00:00.000Z", "field1": "value1"},
			hasField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := TimestampFilter()
			result := filter(tt.input)

			assert.NotNil(t, result)
			assert.Contains(t, result, "@timestamp")

			timestamp, ok := result["@timestamp"].(string)
			assert.True(t, ok)
			assert.NotEmpty(t, timestamp)

			if tt.hasField {
				assert.Equal(t, "2022-01-01T00:00:00.000Z", timestamp)
			} else {
				// Verify the timestamp format is valid
				_, err := time.Parse(timestampFormat, timestamp)
				assert.NoError(t, err)
			}
		})
	}
}
