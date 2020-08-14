package filter

func RemoveFieldFilter(fields []string) FilterFunc {
	return func(m map[string]interface{}) map[string]interface{} {
		for _, field := range fields {
			delete(m, field)
		}
		return m
	}
}
