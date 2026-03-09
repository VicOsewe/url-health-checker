package checker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VicOsewe/url-health-checker/internal/checker"
)

func TestCheck(t *testing.T) {
	// handlers reused across tests
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	hangHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // unblocks when client disconnects or times out
	})

	tests := []struct {
		name            string
		handler         http.Handler
		urlOverride     string
		ctx             func() context.Context
		timeout         time.Duration
		wantStatusCode  int
		wantErr         bool
		wantMinDuration time.Duration
	}{
		{
			name:           "successful 200 response",
			handler:        okHandler,
			ctx:            func() context.Context { return context.Background() },
			timeout:        5 * time.Second,
			wantStatusCode: http.StatusOK,
		},
		{
			name: "404 response is not an error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			ctx:            func() context.Context { return context.Background() },
			timeout:        5 * time.Second,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "500 response is not an error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			ctx:            func() context.Context { return context.Background() },
			timeout:        5 * time.Second,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "invalid URL returns error",
			urlOverride: "not-a-valid-url",
			ctx:         func() context.Context { return context.Background() },
			timeout:     5 * time.Second,
			wantErr:     true,
		},
		{
			name:        "unreachable host returns error",
			urlOverride: "http://localhost:19999",
			ctx:         func() context.Context { return context.Background() },
			timeout:     1 * time.Second,
			wantErr:     true,
		},
		{
			name:    "timeout returns error",
			handler: hangHandler,
			ctx:     func() context.Context { return context.Background() },
			timeout: 100 * time.Millisecond,
			wantErr: true,
		},
		{
			name:    "cancelled context returns error",
			handler: okHandler,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			timeout: 5 * time.Second,
			wantErr: true,
		},
		{
			name:            "duration is recorded",
			handler:         slowHandler,
			ctx:             func() context.Context { return context.Background() },
			timeout:         5 * time.Second,
			wantStatusCode:  http.StatusOK,
			wantMinDuration: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.urlOverride

			// spin up a real local server if the test needs one
			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				defer server.Close()
				url = server.URL
			}

			result := checker.Check(tt.ctx(), url, tt.timeout)

			// URL should always be echoed back
			if result.URL != url {
				t.Errorf("expected URL %q, got %q", url, result.URL)
			}

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("expected an error, got nil (status: %d)", result.StatusCode)
				}
				return // no point checking status code or duration on error
			}

			if result.Error != nil {
				t.Fatalf("unexpected error: %v", result.Error)
			}
			if result.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, result.StatusCode)
			}
			if result.Duration <= 0 {
				t.Errorf("expected positive duration, got %v", result.Duration)
			}
			if tt.wantMinDuration > 0 && result.Duration < tt.wantMinDuration {
				t.Errorf("expected duration >= %v, got %v", tt.wantMinDuration, result.Duration)
			}
		})
	}
}
