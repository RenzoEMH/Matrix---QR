package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apihttp "matrix-qr-apis/go-api/internal/http"
	"matrix-qr-apis/go-api/internal/matrix"
	"matrix-qr-apis/go-api/internal/nodeclient"
)

type stubStats struct {
	resp   *nodeclient.StatsResponse
	err    error
	gotReq nodeclient.StatsRequest
	called bool
}

func (s *stubStats) ComputeStats(_ context.Context, req nodeclient.StatsRequest) (*nodeclient.StatsResponse, error) {
	s.called = true
	s.gotReq = req
	return s.resp, s.err
}

func TestHealth_CORS(t *testing.T) {
	t.Parallel()

	app := apihttp.NewApp(&stubStats{}, testAuth(t), []string{"http://localhost:5173"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("CORS origin = %q", got)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	app := newApp(t, &stubStats{})
	resp := perform(t, app, http.MethodGet, "/health", nil, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestLogin_SuccessAndRejectsBadPassword(t *testing.T) {
	t.Parallel()

	app := newApp(t, &stubStats{})

	ok := perform(t, app, http.MethodPost, "/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`), "")
	defer ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", ok.StatusCode)
	}

	bad := perform(t, app, http.MethodPost, "/auth/login", strings.NewReader(`{"username":"admin","password":"nope"}`), "")
	defer bad.Body.Close()
	if bad.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", bad.StatusCode)
	}
}

func TestProcessMatrix_RequiresJWT(t *testing.T) {
	t.Parallel()

	app := newApp(t, &stubStats{})
	body := strings.NewReader(`{"matrix":[[1,2],[3,4]]}`)

	missing := perform(t, app, http.MethodPost, "/api/v1/matrix", body, "")
	defer missing.Body.Close()
	if missing.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", missing.StatusCode)
	}

	invalid := perform(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3,4]]}`), "not-a-token")
	defer invalid.Body.Close()
	if invalid.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d", invalid.StatusCode)
	}
}

func TestProcessMatrix_Success(t *testing.T) {
	t.Parallel()

	stub := &stubStats{resp: &nodeclient.StatsResponse{
		Max: 6, Min: 1, Average: 3.5, Sum: 21,
		Diagonal: map[string]bool{"q": false, "r": true, "rotated": false},
	}}
	app := newApp(t, stub)
	token := loginToken(t, app)

	body := `{"matrix":[[1,2],[3,4],[5,6]]}`
	resp := perform(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(body), token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
	}

	var got struct {
		Original [][]float64 `json:"original"`
		Rotated  [][]float64 `json:"rotated"`
		QR       struct {
			Q [][]float64 `json:"q"`
			R [][]float64 `json:"r"`
		} `json:"qr"`
		Stats nodeclient.StatsResponse `json:"stats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	wantRotated := matrix.Matrix{{5, 3, 1}, {6, 4, 2}}
	if !matrix.AlmostEqual(got.Rotated, wantRotated, 0) {
		t.Fatalf("rotated = %v, want %v", got.Rotated, wantRotated)
	}
	if got.Stats.Sum != 21 {
		t.Fatalf("stats.sum = %v", got.Stats.Sum)
	}
	if !stub.called {
		t.Fatal("expected stats client to be called")
	}
	if _, ok := stub.gotReq.Matrices["q"]; !ok {
		t.Fatal("stats request missing q")
	}
	if _, ok := stub.gotReq.Matrices["r"]; !ok {
		t.Fatal("stats request missing r")
	}
	if _, ok := stub.gotReq.Matrices["rotated"]; !ok {
		t.Fatal("stats request missing rotated")
	}

	product, err := matrix.MatMul(got.QR.Q, got.QR.R)
	if err != nil {
		t.Fatal(err)
	}
	if !matrix.AlmostEqual(product, got.Original, 1e-8) {
		t.Fatalf("Q R != original")
	}
}

func TestProcessMatrix_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{`},
		{name: "empty matrix", body: `{"matrix":[]}`},
		{name: "jagged", body: `{"matrix":[[1,2],[3]]}`},
		{name: "too large", body: largeMatrixJSON(51)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := newApp(t, &stubStats{})
			token := loginToken(t, app)
			resp := perform(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(tt.body), token)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				raw, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
			}
		})
	}
}

func TestProcessMatrix_StatsUnavailable(t *testing.T) {
	t.Parallel()

	app := newApp(t, &stubStats{err: nodeclient.ErrUnavailable})
	token := loginToken(t, app)
	resp := perform(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3,4]]}`), token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
	}
}

func TestProcessMatrix_WithHTTPNode(t *testing.T) {
	t.Parallel()

	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stats" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"max":4,"min":1,"average":2.5,"sum":10,"diagonal":{"q":false,"r":true,"rotated":false}}`)
	}))
	t.Cleanup(node.Close)

	app := newApp(t, nodeclient.New(node.URL))
	token := loginToken(t, app)
	resp := perform(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3,4]]}`), token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
	}
}

func largeMatrixJSON(n int) string {
	var b bytes.Buffer
	b.WriteString(`{"matrix":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('[')
		for j := 0; j < n; j++ {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('1')
		}
		b.WriteByte(']')
	}
	b.WriteString(`]}`)
	return b.String()
}
