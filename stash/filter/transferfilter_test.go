package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransferFilter(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		field  string
		target string
		expect map[string]interface{}
	}{
		{
			name: "with target",
			input: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"}`,
			},
			field:  "b",
			target: "data",
			expect: map[string]interface{}{
				"a": "aa",
				"data": map[string]interface{}{
					"c": "cc",
				},
			},
		},
		{
			name: "without target",
			input: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"}`,
			},
			field: "b",
			expect: map[string]interface{}{
				"a": "aa",
				"c": "cc",
			},
		},
		{
			name: "without field",
			input: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"}`,
			},
			field: "c",
			expect: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"}`,
			},
		},
		{
			name: "with not json",
			input: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"`,
			},
			field: "b",
			expect: map[string]interface{}{
				"a": "aa",
				"b": `{"c":"cc"`,
			},
		},
		{
			name: "with not string",
			input: map[string]interface{}{
				"a": "aa",
				"b": map[string]interface{}{"c": "cc"},
			},
			field: "b",
			expect: map[string]interface{}{
				"a": "aa",
				"b": map[string]interface{}{"c": "cc"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := TransferFilter(test.field, test.target)(test.input)
			assert.EqualValues(t, test.expect, actual)
		})
	}
}
