package es

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testTime = "2020-09-13T08:22:29.294Z"

func TestBuildIndexFormatter(t *testing.T) {
	tests := []struct {
		name   string
		val    string
		attrs  map[string]interface{}
		expect string
	}{
		{
			name:   "plain text only",
			val:    "yyyy/MM/dd",
			expect: "yyyy/MM/dd",
		},
		{
			name:   "time only",
			val:    "{{yyyy/MM/dd}}",
			expect: time.Now().Format("2006/01/02"),
		},
		{
			name: "attr without time",
			val:  "{.event}",
			attrs: map[string]interface{}{
				"event": "foo",
			},
			expect: "foo",
		},
		{
			name: "attr with time",
			val:  "{.event}-{{yyyy/MM/dd}}",
			attrs: map[string]interface{}{
				"event":      "foo",
				timestampKey: testTime,
			},
			expect: "foo-2020/09/13",
		},
		{
			name: "attr with time, with missing",
			val:  "{.event}-{.foo}-{{yyyy/MM/dd}}",
			attrs: map[string]interface{}{
				"event":      "foo",
				timestampKey: testTime,
			},
			expect: "foo--2020/09/13",
		},
		{
			name: "attr with time, leading alphas",
			val:  "{the.event}-{{yyyy/MM/dd}}",
			attrs: map[string]interface{}{
				"event":      "foo",
				timestampKey: testTime,
			},
			expect: "foo-2020/09/13",
		},
		{
			name: "nested metadata kafka topic",
			val:  "{.@metadata.kafka.topic}-{{yyyy/MM/dd}}",
			attrs: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"kafka": map[string]interface{}{
						"topic": "my-topic",
					},
				},
				timestampKey: testTime,
			},
			expect: "my-topic-2020/09/13",
		},
		{
			name: "nested metadata with missing path",
			val:  "{.@metadata.kafka.missing}-{{yyyy/MM/dd}}",
			attrs: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"kafka": map[string]interface{}{
						"topic": "my-topic",
					},
				},
				timestampKey: testTime,
			},
			expect: "-2020/09/13",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatter := buildIndexFormatter(test.val, time.Local)
			assert.Equal(t, test.expect, formatter(test.attrs))
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name   string
		m      map[string]interface{}
		path   string
		expect string
	}{
		{
			name: "simple field",
			m: map[string]interface{}{
				"field": "value",
			},
			path:   "field",
			expect: "value",
		},
		{
			name: "nested field",
			m: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"kafka": map[string]interface{}{
						"topic": "my-topic",
					},
				},
			},
			path:   "@metadata.kafka.topic",
			expect: "my-topic",
		},
		{
			name: "missing field",
			m: map[string]interface{}{
				"field": "value",
			},
			path:   "missing",
			expect: "",
		},
		{
			name: "missing nested field",
			m: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"kafka": map[string]interface{}{},
				},
			},
			path:   "@metadata.kafka.topic",
			expect: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getNestedValue(test.m, test.path)
			assert.Equal(t, test.expect, result)
		})
	}
}
