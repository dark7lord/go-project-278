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

// ListLinksQuery requests a page of links.
type ListLinksQuery struct {
	Range Range
	Sort  *Sort
}

// ListLinkVisitsQuery requests a page of link visits.
type ListLinkVisitsQuery struct {
	Range Range
	Sort  *Sort
}
