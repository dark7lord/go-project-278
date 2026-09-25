// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"code/internal/application"
)

const fieldID = application.SortFieldID

// pageRange converts an inclusive [start,end] range into a LIMIT/OFFSET pair.
func pageRange(start, end int64) (limit, offset int64) {
	return end - start + 1, start
}
