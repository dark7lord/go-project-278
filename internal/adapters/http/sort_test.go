package httpadapter

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

// testSortFields joins both collections' fields, so one table covers every alias.
var testSortFields = slices.Concat(linksSortFields, visitsSortFields)

func TestParseSortParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    application.Sort
		wantErr error
	}{
		{name: "empty", input: ""},
		{
			name:  "ascending",
			input: `["short_name","ASC"]`,
			want:  application.Sort{Field: "short_name", Asc: true},
		},
		{
			name:  "descending",
			input: `["created_at","DESC"]`,
			want:  application.Sort{Field: "created_at", Asc: false},
		},
		{
			name:  "with surrounding whitespace",
			input: ` ["short_name","ASC"] `,
			want:  application.Sort{Field: "short_name", Asc: true},
		},
		{
			name:  "underscore field",
			input: `["user_agent","ASC"]`,
			want:  application.Sort{Field: "user_agent", Asc: true},
		},
		{
			name:  "short_url sorts as short_name",
			input: `["short_url","DESC"]`,
			want:  application.Sort{Field: "short_name", Asc: false},
		},
		{
			name:  "the dashboard's reffer sorts as referer",
			input: `["reffer","ASC"]`,
			want:  application.Sort{Field: "referer", Asc: true},
		},
		{name: "unknown field", input: `["bogus","ASC"]`, wantErr: ErrSortField},
		{name: "missing quotes", input: `[short_name,ASC]`, wantErr: ErrSortFormat},
		{name: "lowercase direction", input: `["short_name","asc"]`, wantErr: ErrSortFormat},
		{name: "uppercase field", input: `["SHORT_NAME","ASC"]`, wantErr: ErrSortField},
		{name: "single element", input: `["short_name"]`, wantErr: ErrSortFormat},
		{name: "empty brackets", input: `[]`, wantErr: ErrSortFormat},
		{name: "no brackets", input: `short_name,ASC`, wantErr: ErrSortFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSortParam(tt.input, testSortFields)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, got)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
