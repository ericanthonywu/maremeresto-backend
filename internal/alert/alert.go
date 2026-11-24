package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
)

type AlertService struct {
	cfg        *config.Config
	httpClient *http.Client
	mu         sync.Mutex
	lastSent   map[string]time.Time
}

func NewAlertService(cfg *config.Config) *AlertService {
	return &AlertService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		lastSent: make(map[string]time.Time),
	}
}

func (a *AlertService) Enabled() bool {
	return a != nil && a.cfg != nil &&
		strings.TrimSpace(a.cfg.SMTPHost) != "" &&
		strings.TrimSpace(a.cfg.AlertToEmail) != ""
}

func (a *AlertService) WhatsAppEnabled() bool {
	return a != nil && a.cfg != nil &&
		strings.TrimSpace(a.cfg.AlertWhatsAppURL) != "" &&
		strings.TrimSpace(a.cfg.AlertWhatsAppChatID) != ""
}

func (a *AlertService) allowCooldown(key string, cooldown time.Duration) bool {
	if a == nil {
		return false
	}
	if cooldown <= 0 || key == "" {
		return true
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastSent == nil {
		a.lastSent = make(map[string]time.Time)
	}
	now := time.Now()
	if prev, ok := a.lastSent[key]; ok && now.Sub(prev) < cooldown {
		return false
	}
	a.lastSent[key] = now
	return true
}

// SendErrorAlert asynchronously sends an error notification email with trace_id,
// path, error message, and details (e.g. stack trace).
func (a *AlertService) SendErrorAlert(traceID, errMsg, path, details string) {
	if !a.Enabled() && !a.WhatsAppEnabled() {
		return
	}

	subject := fmt.Sprintf("[Backend Alert] Error on %s (Trace ID: %s)", path, traceID)
	body := fmt.Sprintf("Time: %s\nPath: %s\nTrace ID: %s\n\nError Message:\n%s\n\nDetails / Stack Trace:\n%s\n",
		time.Now().Format(time.RFC3339),
		path,
		traceID,
		errMsg,
		details,
	)

	a.dispatchAsync(subject, body, fmt.Sprintf("🚨 *Backend Panic Alert*\nPath: `%s`\nTrace ID: `%s`\nError: %s", path, traceID, errMsg))
}

// SendOperationalAlert asynchronously dispatches an operational alert (e.g. Biteship
// downtime) to configured WhatsApp and Email channels, deduplicated by key within cooldown.
func (a *AlertService) SendOperationalAlert(key, title, details string, cooldown time.Duration) bool {
	if !a.Enabled() && !a.WhatsAppEnabled() {
		return false
	}
	if !a.allowCooldown(key, cooldown) {
		return false
	}

	subject := fmt.Sprintf("[Mareme Alert] %s", title)
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	emailBody := fmt.Sprintf("Time: %s\nTitle: %s\n\n%s\n", now, title, details)
	waMessage := fmt.Sprintf("%s\nWaktu: %s\n\n%s", title, now, details)

	a.dispatchAsync(subject, emailBody, waMessage)
	return true
}

func (a *AlertService) dispatchAsync(subject, emailBody, waMessage string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in async alert sender", "error", r)
			}
		}()

		if a.WhatsAppEnabled() {
			if err := a.sendWhatsApp(context.Background(), waMessage); err != nil {
				slog.Error("failed to send whatsapp alert", "err", err)
			} else {
				slog.Info("whatsapp alert sent successfully", "chat_id", a.cfg.AlertWhatsAppChatID)
			}
		}

		if a.Enabled() {
			if err := a.sendEmail(subject, emailBody); err != nil {
				slog.Error("failed to send error alert email", "subject", subject, "err", err)
			} else {
				slog.Info("error alert email sent successfully", "subject", subject, "to", a.cfg.AlertToEmail)
			}
		}
	}()
}

func (a *AlertService) sendWhatsApp(ctx context.Context, message string) error {
	payload := map[string]string{
		"chatId":  strings.TrimSpace(a.cfg.AlertWhatsAppChatID),
		"message": message,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, strings.TrimSpace(a.cfg.AlertWhatsAppURL), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := a.httpClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("whatsapp bridge returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *AlertService) sendEmail(subject, body string) error {
	from := strings.TrimSpace(a.cfg.AlertFromEmail)
	if from == "" {
		from = a.cfg.SMTPUser
	}
	to := strings.TrimSpace(a.cfg.AlertToEmail)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%s", a.cfg.SMTPHost, a.cfg.SMTPPort)

	var auth smtp.Auth
	if strings.TrimSpace(a.cfg.SMTPUser) != "" {
		auth = smtp.PlainAuth("", a.cfg.SMTPUser, a.cfg.SMTPPassword, a.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
