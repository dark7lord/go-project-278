package application

// Sort requests the ordering of a paginated list by a stored field. The zero
// value, like a field the storage does not know, reads in id order.
type Sort struct {
	Field string
	Asc   bool
}

// PageQuery requests a page of a collection in some order.
type PageQuery struct {
	Range Range
	Sort  Sort
}
