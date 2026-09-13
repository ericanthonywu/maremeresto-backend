package alert

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
)

type AlertService struct {
	cfg *config.Config
}

func NewAlertService(cfg *config.Config) *AlertService {
	return &AlertService{cfg: cfg}
}

func (a *AlertService) Enabled() bool {
	return a != nil && a.cfg != nil &&
		strings.TrimSpace(a.cfg.SMTPHost) != "" &&
		strings.TrimSpace(a.cfg.AlertToEmail) != ""
}

// SendErrorAlert asynchronously sends an error notification email with trace_id,
// path, error message, and details (e.g. stack trace).
func (a *AlertService) SendErrorAlert(traceID, errMsg, path, details string) {
	if !a.Enabled() {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in async email alert sender", "error", r)
			}
		}()

		from := strings.TrimSpace(a.cfg.AlertFromEmail)
		if from == "" {
			from = a.cfg.SMTPUser
		}
		to := strings.TrimSpace(a.cfg.AlertToEmail)

		subject := fmt.Sprintf("[Backend Alert] Error on %s (Trace ID: %s)", path, traceID)
		body := fmt.Sprintf("Time: %s\nPath: %s\nTrace ID: %s\n\nError Message:\n%s\n\nDetails / Stack Trace:\n%s\n",
			time.Now().Format(time.RFC3339),
			path,
			traceID,
			errMsg,
			details,
		)

		msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
			from, to, subject, body)

		addr := fmt.Sprintf("%s:%s", a.cfg.SMTPHost, a.cfg.SMTPPort)

		var auth smtp.Auth
		if strings.TrimSpace(a.cfg.SMTPUser) != "" {
			auth = smtp.PlainAuth("", a.cfg.SMTPUser, a.cfg.SMTPPassword, a.cfg.SMTPHost)
		}

		err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
		if err != nil {
			slog.Error("failed to send error alert email", "trace_id", traceID, "err", err)
		} else {
			slog.Info("error alert email sent successfully", "trace_id", traceID, "to", to)
		}
	}()
}
