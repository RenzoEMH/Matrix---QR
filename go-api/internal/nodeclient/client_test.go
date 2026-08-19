package nodeclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestComputeStats_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v1/stats" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content-type = %s", r.Header.Get("Content-Type"))
		}

		var req StatsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if _, ok := req.Matrices["q"]; !ok {
			t.Error("missing matrices.q")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(StatsResponse{
			Max: 6, Min: 1, Average: 3.5, Sum: 21,
			Diagonal: map[string]bool{"q": false, "r": true, "rotated": false},
		})
	}))
	t.Cleanup(srv.Close)

	got, err := New(srv.URL).ComputeStats(context.Background(), StatsRequest{
		Matrices: map[string][][]float64{"q": {{1}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Max != 6 || got.Min != 1 || got.Sum != 21 {
		t.Fatalf("stats = %+v", got)
	}
}

func TestComputeStats_UnexpectedStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"boom"}`)
	}))
	t.Cleanup(srv.Close)

	_, err := New(srv.URL).ComputeStats(context.Background(), StatsRequest{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want %v", err, ErrUnavailable)
	}
}

func TestComputeStats_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "not-json")
	}))
	t.Cleanup(srv.Close)

	_, err := New(srv.URL).ComputeStats(context.Background(), StatsRequest{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want %v", err, ErrUnavailable)
	}
}

func TestComputeStats_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	_, err := newClient(srv.URL, 30*time.Millisecond).ComputeStats(context.Background(), StatsRequest{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want %v", err, ErrUnavailable)
	}
}

func TestComputeStats_TrimsTrailingSlash(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stats" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"max":1,"min":1,"average":1,"sum":1,"diagonal":{}}`)
	}))
	t.Cleanup(srv.Close)

	if _, err := New(srv.URL+"/").ComputeStats(context.Background(), StatsRequest{}); err != nil {
		t.Fatal(err)
	}
}
