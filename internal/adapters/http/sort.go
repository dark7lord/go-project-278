package httpadapter

import (
	"errors"
	"regexp"

	"code/internal/application"
)

// sortRe matches a sort value like ["short_name","ASC"].
var sortRe = regexp.MustCompile(`^\s*\[\s*"([a-z][a-z0-9_]*)"\s*,\s*"(ASC|DESC)"\s*\]\s*$`)

// linksSortableFields and visitsSortableFields list the fields each collection
// allows sorting on; they mirror the sort dispatch in the repository.
var linksSortableFields = map[application.SortField]struct{}{
	application.SortFieldID:          {},
	application.SortFieldOriginalURL: {},
	application.SortFieldShortName:   {},
}

var visitsSortableFields = map[application.SortField]struct{}{
	application.SortFieldID:        {},
	application.SortFieldLinkID:    {},
	application.SortFieldCreatedAt: {},
	application.SortFieldIP:        {},
	application.SortFieldUserAgent: {},
	application.SortFieldReferer:   {},
	application.SortFieldStatus:    {},
}

var (
	// ErrSortFormat indicates the sort value does not match [field,ASC|DESC].
	ErrSortFormat = errors.New(`invalid sort, expected [field,ASC|DESC]`)
)

// parseSortParam parses a "sort" query parameter value into a sort request;
// an empty value means no sorting.
func parseSortParam(sortParam string) (*application.Sort, error) {
	if sortParam == "" {
		return nil, nil
	}

	matches := sortRe.FindStringSubmatch(sortParam)
	if len(matches) != 3 {
		return nil, ErrSortFormat
	}

	return &application.Sort{
		Field: application.SortField(matches[1]),
		Asc:   matches[2] == "ASC",
	}, nil
}
