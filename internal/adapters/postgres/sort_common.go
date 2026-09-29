package postgres

import (
	"code/internal/application"
)

// sortParams turns a sort request into the page queries' parameters; no sort
// reads in id order. The HTTP layer has already checked the field, and a field
// the query does not know falls back to id order as well.
func sortParams(sort *application.Sort) (field string, asc bool) {
	if sort == nil {
		return string(application.SortFieldID), true
	}

	return string(sort.Field), sort.Asc
}
