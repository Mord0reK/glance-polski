package glance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBeszelWidget_AuthTokenIsFetchedAndReused(t *testing.T) {
	var authCalls atomic.Int32
	var systemsCalls atomic.Int32

	expectedToken := atomic.Value{}
	expectedToken.Store("t1")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/users/auth-with-password":
			authCalls.Add(1)
			_ = r.Body.Close()

			resp := map[string]any{"token": expectedToken.Load().(string)}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		case "/api/collections/systems/records":
			systemsCalls.Add(1)
			got := r.Header.Get("Authorization")
			want := "Bearer " + expectedToken.Load().(string)
			if got != want {
				t.Fatalf("Authorization header mismatch: got %q want %q", got, want)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{
						"id":     "1",
						"name":   "srv",
						"host":   "127.0.0.1",
						"status": "up",
						"info": map[string]any{
							"u":  10,
							"cpu": 0,
							"mp":  0,
							"dp":  0,
						},
					},
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	w := &beszelWidget{
		URL:      srv.URL,
		Identity: "beszel@example.com",
		Password: "secret",
	}
	if err := w.initialize(); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	// First update should login once.
	w.update(context.Background())
	if authCalls.Load() != 1 {
		t.Fatalf("expected 1 auth call, got %d", authCalls.Load())
	}
	if systemsCalls.Load() != 1 {
		t.Fatalf("expected 1 systems call, got %d", systemsCalls.Load())
	}

	// Second update within refresh interval should reuse token.
	w.update(context.Background())
	if authCalls.Load() != 1 {
		t.Fatalf("expected auth call count to remain 1, got %d", authCalls.Load())
	}
	if systemsCalls.Load() != 2 {
		t.Fatalf("expected 2 systems calls, got %d", systemsCalls.Load())
	}
}

func TestBeszelWidget_TokenIsRefreshedAfter3Days(t *testing.T) {
	var authCalls atomic.Int32
	currentToken := atomic.Value{}
	currentToken.Store("t1")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/users/auth-with-password":
			authCalls.Add(1)
			_ = r.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"token": currentToken.Load().(string)})
			return
		case "/api/collections/systems/records":
			got := r.Header.Get("Authorization")
			want := "Bearer " + currentToken.Load().(string)
			if got != want {
				t.Fatalf("Authorization header mismatch: got %q want %q", got, want)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{}})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	bw := &beszelWidget{
		URL:      srv.URL,
		Identity: "beszel@example.com",
		Password: "secret",
	}
	if err := bw.initialize(); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	bw.update(context.Background())
	if authCalls.Load() != 1 {
		t.Fatalf("expected 1 auth call, got %d", authCalls.Load())
	}

	// Force expiry
	bw.tokenFetchedAt = time.Now().Add(-73 * time.Hour)
	currentToken.Store("t2")
	bw.update(context.Background())
	if authCalls.Load() != 2 {
		t.Fatalf("expected 2 auth calls after refresh, got %d", authCalls.Load())
	}
	if bw.Token != "t2" {
		t.Fatalf("expected widget token to be refreshed to t2, got %q", bw.Token)
	}
}

func TestBeszelWidget_FetchChartDataAlsoEnsuresToken(t *testing.T) {
	var authCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/users/auth-with-password":
			authCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "t1"})
			return
		case "/api/collections/system_stats/records":
			if r.Header.Get("Authorization") != "Bearer t1" {
				t.Fatalf("expected Authorization Bearer t1, got %q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{
						"created": "2025-01-01 10:00:00.000Z",
						"stats": map[string]any{"cpu": 1.0, "mp": 2.0, "dp": 3.0, "ns": 4.0, "nr": 5.0},
					},
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	bw := &beszelWidget{URL: srv.URL, Identity: "beszel@example.com", Password: "secret"}
	if err := bw.initialize(); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	_, err := bw.FetchChartData(context.Background(), "sys1", "cpu", "1h")
	if err != nil {
		t.Fatalf("FetchChartData: %v", err)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected 1 auth call, got %d", authCalls.Load())
	}
}
