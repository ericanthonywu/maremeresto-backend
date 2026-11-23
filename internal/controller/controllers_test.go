package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleBiteshipWebhook(t *testing.T) {
	ctrl := NewController(nil, nil, nil)

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{name: "GET ping", method: http.MethodGet, wantStatus: http.StatusOK},
		{name: "HEAD ping", method: http.MethodHead, wantStatus: http.StatusOK},
		{name: "POST empty body", method: http.MethodPost, body: "", wantStatus: http.StatusOK},
		{name: "POST empty json", method: http.MethodPost, body: "{}", wantStatus: http.StatusOK},
		{
			name:       "POST order.status payload",
			method:     http.MethodPost,
			body:       `{"event":"order.status","order_id":"b123","status":"confirmed","courier_waybill_id":"GK-123"}`,
			wantStatus: http.StatusOK,
		},
		{name: "POST invalid json", method: http.MethodPost, body: "{bad", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/webhooks/biteship", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			ctrl.HandleBiteshipWebhook(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("got status %d, want %d (body=%s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}
