package httpapi

import (
	"encoding/json"
	"errors"
	"slices"

	"code/internal/application"
)

// linksSortFields and visitsSortFields list the stored fields a client may
// sort by; they mirror the sort dispatch in the page queries.
var (
	linksSortFields  = []string{"id", "original_url", "short_name"}
	visitsSortFields = []string{
		"id",
		"link_id",
		"created_at",
		"ip",
		"user_agent",
		"referer",
		"status",
	}
)

// sortAliases maps the dashboard's column names to the field they sort as.
var sortAliases = map[string]string{
	// short_url is BASE_URL + "/r/" + short_name: the same order
	"short_url": "short_name",
	// the dashboard's visits column is spelled "reffer"
	"reffer": "referer",
}

var (
	errSortFormat = errors.New(`invalid sort, expected ["field","ASC|DESC"]`)
	errSortField  = errors.New("unsupported sort field")
)

// parseSortParam parses the sort query value ["field","ASC|DESC"] for one of
// fields; an empty value is the zero Sort, id order.
func parseSortParam(sortParam string, fields []string) (application.Sort, error) {
	if sortParam == "" {
		return application.Sort{}, nil
	}

	var pair []string
	if err := json.Unmarshal([]byte(sortParam), &pair); err != nil || len(pair) != 2 {
		return application.Sort{}, errSortFormat
	}
	if pair[1] != "ASC" && pair[1] != "DESC" {
		return application.Sort{}, errSortFormat
	}

	field := pair[0]
	if alias, ok := sortAliases[field]; ok {
		field = alias
	}
	if !slices.Contains(fields, field) {
		return application.Sort{}, errSortField
	}

	return application.Sort{Field: field, Asc: pair[1] == "ASC"}, nil
}
