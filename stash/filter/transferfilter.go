package filter

import (
	jsoniter "github.com/json-iterator/go"
)

func TransferFilter(field, target string) FilterFunc {
	return func(m map[string]interface{}) map[string]interface{} {
		val, ok := m[field]
		if !ok {
			return m
		}

		s, ok := val.(string)
		if !ok {
			return m
		}

		var nm map[string]interface{}
		if err := jsoniter.Unmarshal([]byte(s), &nm); err != nil {
			return m
		}

		delete(m, field)
		if len(target) > 0 {
			m[target] = nm
		} else {
			for k, v := range nm {
				m[k] = v
			}
		}

		return m
	}
}
