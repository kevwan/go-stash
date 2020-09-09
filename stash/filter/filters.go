package filter

import "github.com/tal-tech/go-stash/stash/config"

const (
	filterDrop         = "drop"
	filterRemoveFields = "remove_field"
	opAnd              = "and"
	opOr               = "or"
	typeContains       = "contains"
	typeMatch          = "match"
)

type FilterFunc func(map[string]interface{}) map[string]interface{}

func CreateFilters(p config.Processor) []FilterFunc {
	var filters []FilterFunc

	for _, f := range p.Filters {
		switch f.Action {
		case filterDrop:
			filters = append(filters, DropFilter(f.Conditions))
		case filterRemoveFields:
			filters = append(filters, RemoveFieldFilter(f.Fields))
		}
	}

	return filters
}
