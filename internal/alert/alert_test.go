package alert_test

import (
	"testing"

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
