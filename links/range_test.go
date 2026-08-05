package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRangeParam(t *testing.T) {
	tests := []struct {
		name       string
		rangeParam string
		wantStart  int
		wantEnd    int
		wantErr    error
	}{
		{name: "valid small range", rangeParam: "[0,4]", wantStart: 0, wantEnd: 4},
		{name: "valid middle range", rangeParam: "[5,9]", wantStart: 5, wantEnd: 9},
		{name: "single item", rangeParam: "[3,3]", wantStart: 3, wantEnd: 3},
		{name: "bad format", rangeParam: "invalid", wantErr: ErrRangeFormat},
		{name: "empty brackets", rangeParam: "[]", wantErr: ErrRangeFormat},
		{name: "non-digit start", rangeParam: "[abc,5]", wantErr: ErrRangeFormat},
		{name: "non-digit end", rangeParam: "[5,abc]", wantErr: ErrRangeFormat},
		{name: "negative start", rangeParam: "[-1,5]", wantErr: ErrRangeFormat},
		{name: "overflow start", rangeParam: "[99999999999999999999,5]", wantErr: ErrRangeStart},
		{name: "overflow end", rangeParam: "[5,99999999999999999999]", wantErr: ErrRangeEnd},
		{name: "start > end", rangeParam: "[10,5]", wantErr: ErrRangeNotSatisfiable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseRangeParam(tt.rangeParam)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}
