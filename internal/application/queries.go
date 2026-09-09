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
