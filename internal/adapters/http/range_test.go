package httpadapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRangeParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "bad format", input: "invalid", wantErr: ErrRangeFormat},
		{name: "empty brackets", input: "[]", wantErr: ErrRangeFormat},
		{name: "non-digit start", input: "[abc,5]", wantErr: ErrRangeFormat},
		{name: "non-digit end", input: "[5,abc]", wantErr: ErrRangeFormat},
		{name: "negative start", input: "[-1,5]", wantErr: ErrRangeFormat},
		{name: "overflow start", input: "[99999999999999999999,5]", wantErr: ErrRangeStart},
		{name: "overflow end", input: "[5,99999999999999999999]", wantErr: ErrRangeEnd},
		{name: "start > end", input: "[10,5]", wantErr: ErrRangeNotSatisfiable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseRangeParam(tt.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
