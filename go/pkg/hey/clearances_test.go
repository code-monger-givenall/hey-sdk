package hey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClearancesServiceList(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/clearances" {
			t.Errorf("path = %s, want /clearances", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/html" {
			t.Errorf("Accept = %q, want text/html", got)
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html>
<html><body>
  <span id="email_clearance_999">navigation@example.com</span>
  <article class="card clearance pending" id="clearance_123">
    <span id="name_clearance_123"> Acme &amp; Co. </span>
    <span id="email_clearance_123">sender@example.com</span>
    <span class="clearance__subject">  A useful   subject </span>
    <turbo-frame src="/clearances/entries/456"></turbo-frame>
    <input type="hidden" name="reply_to_topic_id" value="789">
    <form action="/clearances/123" method="post">
      <input type="hidden" name="designation_box_id" value="215744">
      <button data-clearances-target="feedboxButton">Feed</button>
    </form>
    <form action="/clearances/123" method="post">
      <input type="hidden" name="designation_box_id" value="215747">
      <button data-clearances-target="trailboxButton">Paper Trail</button>
    </form>
  </article>
  <article class="clearance" id="clearance_321">
    <span id="name_clearance_321">Second Sender</span>
    <span id="email_clearance_321">second@example.com</span>
    <span class="clearance__subject">Second subject</span>
  </article>
  <article class="clearance" id="not-a-clearance"></article>
</body></html>`))
	})

	got, err := client.Clearances().List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(List) = %d, want 2", len(got))
	}

	wantFirst := PendingClearance{
		ID:           123,
		EntryID:      456,
		TopicID:      789,
		Name:         "Acme & Co.",
		EmailAddress: "sender@example.com",
		Subject:      "A useful subject",
		FeedBoxID:    215744,
		TrailBoxID:   215747,
	}
	if got[0] != wantFirst {
		t.Errorf("first clearance = %#v, want %#v", got[0], wantFirst)
	}
	if got[1].ID != 321 || got[1].EmailAddress != "second@example.com" {
		t.Errorf("second clearance = %#v", got[1])
	}
	if got[1].FeedBoxID != 0 || got[1].TrailBoxID != 0 {
		t.Errorf("second clearance destinations = (%d, %d), want zero values", got[1].FeedBoxID, got[1].TrailBoxID)
	}
}

func TestClearancesServiceUpdates(t *testing.T) {
	tests := []struct {
		name            string
		call            func(*ClearancesService) error
		wantStatus      string
		wantDesignation string
	}{
		{
			name:       "approve to Imbox",
			call:       func(service *ClearancesService) error { return service.Approve(context.Background(), 123, 0) },
			wantStatus: "approved",
		},
		{
			name:            "approve to a designation box",
			call:            func(service *ClearancesService) error { return service.Approve(context.Background(), 123, 215744) },
			wantStatus:      "approved",
			wantDesignation: "215744",
		},
		{
			name:       "deny",
			call:       func(service *ClearancesService) error { return service.Deny(context.Background(), 123) },
			wantStatus: "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("method = %s, want PATCH", r.Method)
				}
				if r.URL.Path != "/clearances/123" {
					t.Errorf("path = %s, want /clearances/123", r.URL.Path)
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
				if got := r.PostForm.Get("status"); got != tt.wantStatus {
					t.Errorf("status = %q, want %q", got, tt.wantStatus)
				}
				if got := r.PostForm.Get("designation_box_id"); got != tt.wantDesignation {
					t.Errorf("designation_box_id = %q, want %q", got, tt.wantDesignation)
				}
				w.WriteHeader(http.StatusNoContent)
			})

			if err := tt.call(client.Clearances()); err != nil {
				t.Fatalf("update: %v", err)
			}
		})
	}
}

func TestClearancesServiceUpdateAcceptsRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/clearances")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	client := NewClient(
		&Config{BaseURL: server.URL},
		&StaticTokenProvider{Token: "test-token"},
		WithHTTPClient(httpClient),
		WithMaxRetries(0),
		WithBaseDelay(time.Millisecond),
		WithMaxJitter(time.Millisecond),
	)

	if err := client.Clearances().Deny(context.Background(), 123); err != nil {
		t.Fatalf("Deny: %v", err)
	}
}

func TestClearancesServiceRejectsInvalidIDsBeforeRequest(t *testing.T) {
	var requests atomic.Int64
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.Clearances().Deny(context.Background(), 0); err == nil || !strings.Contains(err.Error(), "positive") {
		t.Fatalf("Deny(0) error = %v, want positive ID error", err)
	}
	if err := client.Clearances().Approve(context.Background(), 123, -1); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("Approve(-1) error = %v, want negative box ID error", err)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
}

func TestClientClearancesReturnsCachedService(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	first := client.Clearances()
	second := client.Clearances()
	if first == nil || first != second {
		t.Fatalf("Clearances service pointers = (%p, %p), want same non-nil service", first, second)
	}
}
