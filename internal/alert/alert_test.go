package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/alert"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
)

func TestAlertServiceEnabled(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "",
		AlertToEmail: "admin@example.com",
	}
	svc := alert.NewAlertService(cfg)

	if svc.Enabled() {
		t.Errorf("expected alert service to be disabled when SMTPHost is empty")
	}

	cfg.SMTPHost = "smtp.gmail.com"
	if !svc.Enabled() {
		t.Errorf("expected alert service to be enabled when SMTPHost and AlertToEmail are provided")
	}
}

func TestSendErrorAlertWhenDisabled(t *testing.T) {
	cfg := &config.Config{}
	svc := alert.NewAlertService(cfg)

	// Should not panic or attempt to send when disabled
	svc.SendErrorAlert("trace-123", "test error", "/api/v1/test", "stack trace info")
}

func TestSendOperationalAlertWhatsAppAndCooldown(t *testing.T) {
	var calls atomic.Int32
	received := make(chan map[string]string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			received <- body
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := alert.NewAlertService(&config.Config{
		AlertWhatsAppURL:    srv.URL,
		AlertWhatsAppChatID: "120363430827193538@g.us",
	})
	if !svc.WhatsAppEnabled() {
		t.Fatalf("expected WhatsAppEnabled to be true")
	}

	if !svc.SendOperationalAlert("biteship:quote_failed", "Biteship Down", "timeout", 5*time.Minute) {
		t.Fatalf("expected first alert to be dispatched")
	}
	if svc.SendOperationalAlert("biteship:quote_failed", "Biteship Down", "timeout", 5*time.Minute) {
		t.Fatalf("expected duplicate alert within cooldown to be suppressed")
	}

	select {
	case payload := <-received:
		if payload["chatId"] != "120363430827193538@g.us" {
			t.Fatalf("unexpected chatId: %q", payload["chatId"])
		}
		if payload["message"] == "" {
			t.Fatalf("expected non-empty alert message")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for WhatsApp alert HTTP POST")
	}

	time.Sleep(50 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", got)
	}
}
