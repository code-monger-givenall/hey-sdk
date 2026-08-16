package hey

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTopicAndEntryStatusMutations(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		wantPath   string
		call       func(*Client) error
	}{
		{
			name:       "restore topic",
			wantMethod: http.MethodPut,
			wantPath:   "/topics/123/status/active",
			call: func(client *Client) error {
				return client.Topics().Restore(context.Background(), 123)
			},
		},
		{
			name:       "mark entry spam",
			wantMethod: http.MethodPut,
			wantPath:   "/entries/456/status/spam",
			call: func(client *Client) error {
				return client.Entries().MarkSpam(context.Background(), 456)
			},
		},
		{
			name:       "cancel bubble up",
			wantMethod: http.MethodDelete,
			wantPath:   "/topics/123/bubble_up",
			call: func(client *Client) error {
				return client.Topics().CancelBubbleUp(context.Background(), 123)
			},
		},
		{
			name:       "bubble up now",
			wantMethod: http.MethodPost,
			wantPath:   "/topics/123/bubble_up_now",
			call: func(client *Client) error {
				return client.Topics().BubbleUpNow(context.Background(), 123)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod {
					t.Errorf("method = %s, want %s", r.Method, tt.wantMethod)
				}
				if r.URL.Path != tt.wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, tt.wantPath)
				}
				w.WriteHeader(http.StatusNoContent)
			})

			if err := tt.call(client); err != nil {
				t.Fatalf("mutation: %v", err)
			}
		})
	}
}

func TestTopicsServiceScheduleBubbleUp(t *testing.T) {
	tests := []struct {
		name         string
		waitingOn    bool
		wantRawQuery string
	}{
		{name: "standard reminder"},
		{name: "waiting on", waitingOn: true, wantRawQuery: "waiting_on=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/topics/123/bubble_up" {
					t.Errorf("path = %s, want /topics/123/bubble_up", r.URL.Path)
				}
				if r.URL.RawQuery != tt.wantRawQuery {
					t.Errorf("query = %q, want %q", r.URL.RawQuery, tt.wantRawQuery)
				}
				if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
					t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", got)
				}
				if got := r.Header.Get("Accept"); got != "*/*" {
					t.Errorf("Accept = %q, want */*", got)
				}
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm: %v", err)
				}
				if got := r.PostForm.Get("date"); got != "2026-08-20" {
					t.Errorf("date = %q, want 2026-08-20", got)
				}
				w.WriteHeader(http.StatusNoContent)
			})

			if err := client.Topics().ScheduleBubbleUp(context.Background(), 123, "2026-08-20", tt.waitingOn); err != nil {
				t.Fatalf("ScheduleBubbleUp: %v", err)
			}
		})
	}
}

func TestTopicControlsRejectInvalidInputBeforeRequest(t *testing.T) {
	var requests atomic.Int64
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name    string
		call    func() error
		message string
	}{
		{name: "restore ID", call: func() error { return client.Topics().Restore(context.Background(), 0) }, message: "topic ID must be positive"},
		{name: "spam ID", call: func() error { return client.Entries().MarkSpam(context.Background(), 0) }, message: "entry ID must be positive"},
		{name: "schedule ID", call: func() error { return client.Topics().ScheduleBubbleUp(context.Background(), 0, "2026-08-20", false) }, message: "topic ID must be positive"},
		{name: "schedule date", call: func() error { return client.Topics().ScheduleBubbleUp(context.Background(), 123, "August 20", false) }, message: "YYYY-MM-DD"},
		{name: "cancel ID", call: func() error { return client.Topics().CancelBubbleUp(context.Background(), 0) }, message: "topic ID must be positive"},
		{name: "now ID", call: func() error { return client.Topics().BubbleUpNow(context.Background(), 0) }, message: "topic ID must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v, want %q", err, tt.message)
			}
		})
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
}

func TestCheckMutationResponseAcceptsRedirects(t *testing.T) {
	for _, status := range []int{http.StatusFound, http.StatusSeeOther} {
		if err := checkMutationResponse(&http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header)}); err != nil {
			t.Errorf("status %d: %v", status, err)
		}
	}

	response := &http.Response{StatusCode: http.StatusBadRequest, Status: http.StatusText(http.StatusBadRequest), Header: make(http.Header)}
	if err := checkMutationResponse(response); err == nil {
		t.Fatal("expected non-success response error")
	}
}
