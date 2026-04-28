package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewNotifier_EmptyURL(t *testing.T) {
	_, err := NewNotifier(WebhookConfig{})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewNotifier_ValidURL(t *testing.T) {
	n, err := NewNotifier(WebhookConfig{URL: "http://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}

func TestNotifier_Send_Success(t *testing.T) {
	var received WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n, _ := NewNotifier(WebhookConfig{URL: server.URL, Timeout: 2 * time.Second})
	d := Diff{
		Added:   []Route{{Destination: "10.0.0.0/8"}},
		Removed: []Route{{Destination: "192.168.1.0/24"}},
	}
	if err := n.Send("testhost", d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received.Added) != 1 || received.Added[0] != "10.0.0.0/8" {
		t.Errorf("unexpected added: %v", received.Added)
	}
	if len(received.Removed) != 1 || received.Removed[0] != "192.168.1.0/24" {
		t.Errorf("unexpected removed: %v", received.Removed)
	}
	if received.Host != "testhost" {
		t.Errorf("unexpected host: %s", received.Host)
	}
}

func TestNotifier_Send_CustomHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-Api-Key"); v != "secret" {
			t.Errorf("expected header X-Api-Key=secret, got %q", v)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	n, _ := NewNotifier(WebhookConfig{
		URL:     server.URL,
		Headers: map[string]string{"X-Api-Key": "secret"},
	})
	if err := n.Send("host", Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n, _ := NewNotifier(WebhookConfig{URL: server.URL})
	if err := n.Send("host", Diff{}); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestNotifier_Send_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n, _ := NewNotifier(WebhookConfig{URL: server.URL, Timeout: 50 * time.Millisecond})
	if err := n.Send("host", Diff{}); err == nil {
		t.Fatal("expected error due to client timeout")
	}
}
