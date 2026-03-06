package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TelegramNotifier struct {
	botToken     string
	chatID       string
	serverName   string
	timeout      time.Duration
	dedupeWindow time.Duration
	baseURL      string
	httpClient   *http.Client

	mu       sync.Mutex
	lastSent map[string]time.Time
}

type telegramSendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramSendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func NewTelegramNotifier(botToken, chatID, serverName string, timeout, dedupeWindow time.Duration) *TelegramNotifier {
	name := strings.TrimSpace(serverName)
	if name == "" {
		name = "idsai"
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if dedupeWindow < 0 {
		dedupeWindow = 0
	}

	return &TelegramNotifier{
		botToken:     strings.TrimSpace(botToken),
		chatID:       strings.TrimSpace(chatID),
		serverName:   name,
		timeout:      timeout,
		dedupeWindow: dedupeWindow,
		baseURL:      "https://api.telegram.org",
		httpClient:   &http.Client{Timeout: timeout},
		lastSent:     make(map[string]time.Time),
	}
}

func (n *TelegramNotifier) Enabled() bool {
	return n != nil && n.botToken != "" && n.chatID != ""
}

func (n *TelegramNotifier) NotifyStarted(ctx context.Context) error {
	if !n.Enabled() {
		return nil
	}
	msg := fmt.Sprintf("🟢 Server started\nService: %s\nTime: %s", n.serverName, n.now())
	return n.sendWithDedupe(ctx, "server.started", msg)
}

func (n *TelegramNotifier) NotifyCritical(ctx context.Context, err error) error {
	if !n.Enabled() {
		return nil
	}
	errText := trimMessage(errorText(err), 900)
	msg := fmt.Sprintf("🔴 Critical server error\nService: %s\nError: %s\nTime: %s", n.serverName, errText, n.now())
	key := "server.critical:" + compactKey(errText)
	return n.sendWithDedupe(ctx, key, msg)
}

func (n *TelegramNotifier) NotifyUnhandledPanic(ctx context.Context, recovered any, stack []byte) error {
	if !n.Enabled() {
		return nil
	}
	panicText := fmt.Sprintf("%v", recovered)
	panicText = trimMessage(strings.TrimSpace(panicText), 500)
	stackText := trimMessage(strings.TrimSpace(string(stack)), 1000)
	msg := fmt.Sprintf("🔴 Unhandled panic\nService: %s\nError: %s\nStack: %s\nTime: %s", n.serverName, panicText, stackText, n.now())
	key := "server.panic:" + compactKey(panicText)
	return n.sendWithDedupe(ctx, key, msg)
}

func (n *TelegramNotifier) NotifyDBFailure(ctx context.Context, err error) error {
	if !n.Enabled() {
		return nil
	}
	errText := trimMessage(errorText(err), 900)
	msg := fmt.Sprintf("🔴 Database connection failure\nService: %s\nError: %s\nTime: %s", n.serverName, errText, n.now())
	key := "db.failure:" + compactKey(errText)
	return n.sendWithDedupe(ctx, key, msg)
}

func (n *TelegramNotifier) NotifyRecovered(ctx context.Context, details string) error {
	if !n.Enabled() {
		return nil
	}
	details = strings.TrimSpace(details)
	if details == "" {
		details = "service healthy"
	}
	details = trimMessage(details, 320)
	msg := fmt.Sprintf("🟢 Server recovered\nService: %s\nDetails: %s\nTime: %s", n.serverName, details, n.now())
	key := "server.recovered:" + compactKey(details)
	return n.sendWithDedupe(ctx, key, msg)
}

func (n *TelegramNotifier) NotifyHeartbeat(ctx context.Context) error {
	if !n.Enabled() {
		return nil
	}
	msg := fmt.Sprintf("💓 Server is alive\nService: %s\nTime: %s", n.serverName, n.now())
	return n.sendDirect(ctx, msg)
}

func (n *TelegramNotifier) NotifyStopped(ctx context.Context, reason string) error {
	if !n.Enabled() {
		return nil
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "graceful shutdown"
	}
	reason = trimMessage(reason, 320)
	msg := fmt.Sprintf("🟡 Server stopped\nService: %s\nReason: %s\nTime: %s", n.serverName, reason, n.now())
	key := "server.stopped:" + compactKey(reason)
	return n.sendWithDedupe(ctx, key, msg)
}

func (n *TelegramNotifier) sendWithDedupe(ctx context.Context, dedupeKey, text string) error {
	if !n.Enabled() {
		return nil
	}
	if !n.shouldSend(strings.TrimSpace(dedupeKey)) {
		return nil
	}
	return n.sendDirect(ctx, text)
}

func (n *TelegramNotifier) sendDirect(ctx context.Context, text string) error {
	if !n.Enabled() {
		return nil
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("telegram message is empty")
	}

	payload := telegramSendMessageRequest{
		ChatID:                n.chatID,
		Text:                  text,
		DisableWebPagePreview: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(n.baseURL, "/"), n.botToken)
	requestCtx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram http status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var apiResp telegramSendMessageResponse
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &apiResp); err == nil && !apiResp.OK {
			if apiResp.Description == "" {
				return errors.New("telegram api returned not ok")
			}
			return errors.New(apiResp.Description)
		}
	}

	return nil
}

func (n *TelegramNotifier) shouldSend(key string) bool {
	if key == "" || n.dedupeWindow <= 0 {
		return true
	}
	now := time.Now()

	n.mu.Lock()
	defer n.mu.Unlock()

	if last, ok := n.lastSent[key]; ok && now.Sub(last) < n.dedupeWindow {
		return false
	}
	n.lastSent[key] = now

	if len(n.lastSent) > 256 {
		for k, ts := range n.lastSent {
			if now.Sub(ts) > n.dedupeWindow*2 {
				delete(n.lastSent, k)
			}
		}
	}
	return true
}

func (n *TelegramNotifier) now() string {
	return time.Now().Format(time.RFC3339)
}

func errorText(err error) string {
	if err == nil {
		return "unknown"
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "unknown"
	}
	return msg
}

func compactKey(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func trimMessage(v string, limit int) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "-"
	}
	if limit <= 0 || len(v) <= limit {
		return v
	}
	if limit < 4 {
		return v[:limit]
	}
	return v[:limit-3] + "..."
}
