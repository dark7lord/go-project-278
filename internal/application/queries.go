package application

// SortField identifies a sortable column on a paginated collection.
type SortField string

// Sortable field names shared by the HTTP and PostgreSQL layers.
const (
	SortFieldID          SortField = "id"
	SortFieldShortName   SortField = "short_name"
	SortFieldOriginalURL SortField = "original_url"
	SortFieldLinkID      SortField = "link_id"
	SortFieldCreatedAt   SortField = "created_at"
	SortFieldIP          SortField = "ip"
	SortFieldUserAgent   SortField = "user_agent"
	SortFieldReferer     SortField = "referer"
	SortFieldStatus      SortField = "status"
)

// Sort requests the ordering of a paginated list.
type Sort struct {
	Field SortField
	Asc   bool
}

// ListLinksQuery requests a paginated list of links.
type ListLinksQuery struct {
	Start int64
	End   int64
	Sort  *Sort
}

// ListLinkVisitsQuery requests a paginated list of link visits.
type ListLinkVisitsQuery struct {
	Start int64
	End   int64
	Sort  *Sort
}

// RangePage is the result of a paginated range query.
type RangePage[T any] struct {
	Items []T
	Start int64
	Total int64
}
