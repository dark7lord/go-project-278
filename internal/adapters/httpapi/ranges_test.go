package httpapi

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestCapRange(t *testing.T) {
	tests := []struct {
		name string
		in   application.Range
		want application.Range
	}{
		{name: "within a page", in: application.Range{First: 0, Last: 9}, want: application.Range{First: 0, Last: 9}},
		{name: "a full page", in: application.Range{First: 0, Last: 999}, want: application.Range{First: 0, Last: 999}},
		{name: "wider", in: application.Range{First: 5, Last: 5000}, want: application.Range{First: 5, Last: 1004}},
		{name: "open", in: application.Range{First: 5, Last: math.MaxInt64}, want: application.Range{First: 5, Last: 1004}},
		{
			name: "open near the int64 limit",
			in:   application.Range{First: math.MaxInt64 - 5, Last: math.MaxInt64},
			want: application.Range{First: math.MaxInt64 - 5, Last: math.MaxInt64},
		},
		{
			name: "short suffix",
			in:   application.Range{Suffix: true, Length: 20},
			want: application.Range{Suffix: true, Length: 20},
		},
		{
			name: "long suffix",
			in:   application.Range{Suffix: true, Length: 5000},
			want: application.Range{Suffix: true, Length: maxPageSize},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, capRange(tt.in))
		})
	}
}

func TestRequestRange(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		header         string
		want           application.Range
		wantFromHeader bool
		wantErr        error
	}{
		{name: "neither reads the first page", target: "/links", want: application.Range{First: 0, Last: 999}},
		{
			name:           "honoured header",
			target:         "/links",
			header:         "links=-20",
			want:           application.Range{Suffix: true, Length: 20},
			wantFromHeader: true,
		},
		{
			name:   "ignored header reads the first page",
			target: "/links",
			header: "bytes=0-9",
			want:   application.Range{First: 0, Last: 999},
		},
		{
			name:   "query wins over header",
			target: "/links?range=[5,9]",
			header: "links=0-1",
			want:   application.Range{First: 5, Last: 9},
		},
		{name: "malformed query is an error", target: "/links?range=[9,5]", wantErr: errRangeInverted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, tt.target, nil)
			if tt.header != "" {
				c.Request.Header.Set("Range", tt.header)
			}

			got, fromHeader, err := requestRange(c, linksUnit)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantFromHeader, fromHeader)
		})
	}
}

func TestParseRangeParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    application.Range
		wantErr error
	}{
		{name: "valid", input: "[5,9]", want: application.Range{First: 5, Last: 9}},
		{name: "valid with spaces", input: "[0, 4]", want: application.Range{First: 0, Last: 4}},
		{name: "valid with surrounding whitespace", input: " [0,4] ", want: application.Range{First: 0, Last: 4}},
		{name: "wider than a page", input: "[0,5000]", want: application.Range{First: 0, Last: 5000}},
		{name: "bad format", input: "invalid", wantErr: errRangeFormat},
		{name: "empty brackets", input: "[]", wantErr: errRangeFormat},
		{name: "non-digit start", input: "[abc,5]", wantErr: errRangeFormat},
		{name: "non-digit end", input: "[5,abc]", wantErr: errRangeFormat},
		{name: "negative start", input: "[-1,5]", wantErr: errRangeFormat},
		{name: "prefix junk", input: "abc[0,4]", wantErr: errRangeFormat},
		{name: "suffix junk", input: "[0,4]junk", wantErr: errRangeFormat},
		{name: "trailing text after space", input: "[0, 4] x", wantErr: errRangeFormat},
		{name: "RFC form is not a query value", input: "0-4", wantErr: errRangeFormat},
		{name: "overflow start", input: "[99999999999999999999,5]", wantErr: errRangeFormat},
		{name: "overflow end", input: "[5,99999999999999999999]", wantErr: errRangeFormat},
		{name: "fraction", input: "[0.5,5]", wantErr: errRangeFormat},
		{name: "three bounds", input: "[0,4,9]", wantErr: errRangeFormat},
		{name: "start > end", input: "[10,5]", wantErr: errRangeInverted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRangeParam(tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

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
