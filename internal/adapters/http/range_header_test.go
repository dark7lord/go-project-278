package httpadapter

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"code/internal/application"
)

func TestParseRangeHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   application.Range
		wantOK bool
	}{
		{name: "closed", header: "links=0-9", want: application.Range{First: 0, Last: 9}, wantOK: true},
		{name: "unit is case-insensitive", header: "Links=5-9", want: application.Range{First: 5, Last: 9}, wantOK: true},
		{name: "single item", header: "links=3-3", want: application.Range{First: 3, Last: 3}, wantOK: true},
		{name: "open", header: "links=10-", want: application.Range{First: 10, Last: math.MaxInt64}, wantOK: true},
		{name: "suffix", header: "links=-20", want: application.Range{Suffix: true, Length: 20}, wantOK: true},
		{name: "zero suffix", header: "links=-0", want: application.Range{Suffix: true}, wantOK: true},
		{name: "absent", header: ""},
		{name: "other unit", header: "bytes=0-9"},
		{name: "another collection", header: "link_visits=0-9"},
		{name: "no unit", header: "0-9"},
		{name: "bracket form", header: "[0,9]"},
		{name: "several ranges", header: "links=0-9,20-29"},
		{name: "empty set", header: "links="},
		{name: "dash only", header: "links=-"},
		{name: "last before first", header: "links=9-5"},
		{name: "spaces", header: "links= 0-9"},
		{name: "not digits", header: "links=a-b"},
		{name: "negative first", header: "links=--5"},
		{name: "overflow first", header: "links=99999999999999999999-"},
		{name: "overflow last", header: "links=0-99999999999999999999"},
		{name: "overflow suffix", header: "links=-99999999999999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseRangeHeader(tt.header, linksUnit)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
