package httpadapter

import (
	"errors"
	"regexp"

	"code/internal/application"
)

// sortRe matches a sort value like ["short_name","ASC"].
var sortRe = regexp.MustCompile(`^\s*\[\s*"([a-z][a-z0-9_]*)"\s*,\s*"(ASC|DESC)"\s*\]\s*$`)

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
	// ErrSortFormat indicates the sort value does not match [field,ASC|DESC].
	ErrSortFormat = errors.New(`invalid sort, expected [field,ASC|DESC]`)
)

// parseSortParam parses a "sort" query parameter value into a sort request on
// one of fields; an empty value means no sorting.
func parseSortParam(sortParam string, fields map[string]application.SortField) (*application.Sort, error) {
	if sortParam == "" {
		return nil, nil
	}

	matches := sortRe.FindStringSubmatch(sortParam)
	if len(matches) != 3 {
		return nil, ErrSortFormat
	}

	field, ok := fields[matches[1]]
	if !ok {
		return nil, application.ErrSortField
	}

	return &application.Sort{Field: field, Asc: matches[2] == "ASC"}, nil
}
