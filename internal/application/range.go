package application

// Range selects items of a collection by position, like an RFC 9110 range:
// the inclusive span [First, Last] or, when Suffix is set, the last Length items.
type Range struct {
	First  int64
	Last   int64
	Suffix bool
	Length int64
}

// Resolve places the range on a collection of total items and returns the
// inclusive span to read; ok is false when no item falls into the range.
func (r Range) Resolve(total int64) (first, last int64, ok bool) {
	if r.Suffix {
		if r.Length == 0 || total == 0 {
			return 0, 0, false
		}

		return max(total-r.Length, 0), total - 1, true
	}

	if r.First >= total {
		return 0, 0, false
	}

	return r.First, min(r.Last, total-1), true
}

// RangePage is the result of a paginated range query: Items start at position
// First of Total; no Items in a non-empty collection means the range missed it.
type RangePage[T any] struct {
	Items []T
	First int64
	Total int64
}
