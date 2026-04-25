package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookConfig holds configuration for webhook notifications.
type WebhookConfig struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
}

// WebhookPayload is the JSON body sent to the webhook endpoint.
type WebhookPayload struct {
	Timestamp string   `json:"timestamp"`
	Host      string   `json:"host"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
}

// Notifier sends diff notifications to a webhook.
type Notifier struct {
	cfg    WebhookConfig
	client *http.Client
}

// NewNotifier creates a Notifier with the given config.
func NewNotifier(cfg WebhookConfig) (*Notifier, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook URL must not be empty")
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Notifier{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

// Send posts the diff as a JSON payload to the configured webhook URL.
func (n *Notifier) Send(host string, d Diff) error {
	added := make([]string, len(d.Added))
	for i, r := range d.Added {
		added[i] = r.Destination
	}
	removed := make([]string, len(d.Removed))
	for i, r := range d.Removed {
		removed[i] = r.Destination
	}
	payload := WebhookPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Host:      host,
		Added:     added,
		Removed:   removed,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, n.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.cfg.Headers {
		req.Header.Set(k, v)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
	}
	return nil
}
