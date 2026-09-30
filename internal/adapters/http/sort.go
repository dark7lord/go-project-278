package httpadapter

import (
	"encoding/json"
	"errors"

	"code/internal/application"
)

// linksSortFields and visitsSortFields map the names a client may sort by to
// the stored fields; they mirror the sort dispatch in the repository.
var linksSortFields = map[string]application.SortField{
	"id":           application.SortFieldID,
	"original_url": application.SortFieldOriginalURL,
	"short_name":   application.SortFieldShortName,
	// short_url is BASE_URL + "/r/" + short_name: the same order
	"short_url": application.SortFieldShortName,
}

var visitsSortFields = map[string]application.SortField{
	"id":         application.SortFieldID,
	"link_id":    application.SortFieldLinkID,
	"created_at": application.SortFieldCreatedAt,
	"ip":         application.SortFieldIP,
	"user_agent": application.SortFieldUserAgent,
	"referer":    application.SortFieldReferer,
	"status":     application.SortFieldStatus,
	// the dashboard's visits column is spelled "reffer"
	"reffer": application.SortFieldReferer,
}

var (
	// ErrSortFormat indicates the sort value is not a JSON array ["field","ASC|DESC"].
	ErrSortFormat = errors.New(`invalid sort, expected [field,ASC|DESC]`)
	// ErrSortField indicates an unsupported sort field.
	ErrSortField = errors.New("unsupported sort field")
)

// parseSortParam parses a "sort" query parameter value into a sort request on
// one of fields; an empty value means no sorting.
func parseSortParam(sortParam string, fields map[string]application.SortField) (*application.Sort, error) {
	if sortParam == "" {
		return nil, nil
	}

	var pair []string
	if err := json.Unmarshal([]byte(sortParam), &pair); err != nil || len(pair) != 2 {
		return nil, ErrSortFormat
	}
	if pair[1] != "ASC" && pair[1] != "DESC" {
		return nil, ErrSortFormat
	}

	field, ok := fields[pair[0]]
	if !ok {
		return nil, ErrSortField
	}

	return &application.Sort{Field: field, Asc: pair[1] == "ASC"}, nil
}
