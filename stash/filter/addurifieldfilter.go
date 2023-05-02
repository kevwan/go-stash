package filter

import (
	"encoding/hex"
	"strings"
)

func AddUriFieldFilter(inField, outField string) FilterFunc {
	return func(m map[string]interface{}) map[string]interface{} {
		if val, ok := m[inField].(string); ok {
			var datas []string
			idx := strings.Index(val, "?")
			if idx < 0 {
				datas = strings.Split(val, "/")
			} else {
				datas = strings.Split(val[:idx], "/")
			}

			for i, data := range datas {
				if IsObjectIdHex(data) {
					datas[i] = "*"
				}
			}

			m[outField] = strings.Join(datas, "/")
		}

		return m
	}
}

// IsObjectIdHex returns whether s is a valid hex representation of
// an ObjectId. See the ObjectIdHex function.
func IsObjectIdHex(s string) bool {
	if len(s) != 24 {
		return false
	}

	_, err := hex.DecodeString(s)
	return err == nil
}
