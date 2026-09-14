package application

// ListLinksQuery requests a paginated list of links.
type ListLinksQuery struct {
	Start int64
	End   int64
}

// ListLinkVisitsQuery requests a paginated list of link visits.
type ListLinkVisitsQuery struct {
	Start int64
	End   int64
}

// RangePage is the result of a paginated range query.
type RangePage[T any] struct {
	Items []T
	Start int64
	Total int64
}
