package httpadapter

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
		{name: "malformed query is an error", target: "/links?range=[9,5]", wantErr: ErrRangeInverted},
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
