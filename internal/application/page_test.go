package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeResolve(t *testing.T) {
	tests := []struct {
		name      string
		r         Range
		total     int64
		wantFirst int64
		wantLast  int64
		wantOK    bool
	}{
		{name: "inside", r: Range{First: 2, Last: 4}, total: 10, wantFirst: 2, wantLast: 4, wantOK: true},
		{name: "last cut to the end", r: Range{First: 8, Last: 20}, total: 10, wantFirst: 8, wantLast: 9, wantOK: true},
		{name: "last item", r: Range{First: 9, Last: 9}, total: 10, wantFirst: 9, wantLast: 9, wantOK: true},
		{name: "first at the end", r: Range{First: 10, Last: 20}, total: 10},
		{name: "first past the end", r: Range{First: 50, Last: 60}, total: 10},
		{name: "empty collection", r: Range{First: 0, Last: 9}, total: 0},
		{name: "suffix", r: Range{Suffix: true, Length: 3}, total: 10, wantFirst: 7, wantLast: 9, wantOK: true},
		{name: "suffix longer than total", r: Range{Suffix: true, Length: 50}, total: 10, wantLast: 9, wantOK: true},
		{name: "zero suffix", r: Range{Suffix: true, Length: 0}, total: 10},
		{name: "suffix on empty collection", r: Range{Suffix: true, Length: 3}, total: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, last, ok := tt.r.Resolve(tt.total)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantFirst, first)
			assert.Equal(t, tt.wantLast, last)
		})
	}
}
