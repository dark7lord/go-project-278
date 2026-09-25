package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkErrors(t *testing.T) {
	td := setupTestDB(t)

	const missingLinkPath = "/api/links/999"

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "GetLink / not found",
			method:     http.MethodGet,
			path:       missingLinkPath,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "UpdateLink / not found",
			method:     http.MethodPut,
			path:       missingLinkPath,
			body:       `{"original_url": "https://nowhere.com", "short_name": "ghost"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "DeleteLink / not found",
			method:     http.MethodDelete,
			path:       missingLinkPath,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			w := performRequest(t, tx.router, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.wantStatus, w.Code)
			assertErrorBody(t, w)
		})
	}
}
