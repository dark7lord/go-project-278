package httpadapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

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
		{name: "bad format", input: "invalid", wantErr: ErrRangeFormat},
		{name: "empty brackets", input: "[]", wantErr: ErrRangeFormat},
		{name: "non-digit start", input: "[abc,5]", wantErr: ErrRangeFormat},
		{name: "non-digit end", input: "[5,abc]", wantErr: ErrRangeFormat},
		{name: "negative start", input: "[-1,5]", wantErr: ErrRangeFormat},
		{name: "prefix junk", input: "abc[0,4]", wantErr: ErrRangeFormat},
		{name: "suffix junk", input: "[0,4]junk", wantErr: ErrRangeFormat},
		{name: "trailing text after space", input: "[0, 4] x", wantErr: ErrRangeFormat},
		{name: "RFC form is not a query value", input: "0-4", wantErr: ErrRangeFormat},
		{name: "overflow start", input: "[99999999999999999999,5]", wantErr: ErrRangeStart},
		{name: "overflow end", input: "[5,99999999999999999999]", wantErr: ErrRangeEnd},
		{name: "start > end", input: "[10,5]", wantErr: ErrRangeInverted},
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
