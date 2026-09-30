package application

// Sort requests the ordering of a paginated list by a stored field. The zero
// value, like a field the storage does not know, reads in id order.
type Sort struct {
	Field string
	Asc   bool
}

// ListLinksQuery requests a page of links.
type ListLinksQuery struct {
	Range Range
	Sort  Sort
}

// ListLinkVisitsQuery requests a page of link visits.
type ListLinkVisitsQuery struct {
	Range Range
	Sort  Sort
}
