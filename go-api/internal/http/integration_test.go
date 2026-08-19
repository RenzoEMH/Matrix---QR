package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"matrix-qr-apis/go-api/internal/matrix"
	"matrix-qr-apis/go-api/internal/nodeclient"
)

type matrixAPIResponse struct {
	Original [][]float64 `json:"original"`
	Rotated  [][]float64 `json:"rotated"`
	QR       struct {
		Q [][]float64 `json:"q"`
		R [][]float64 `json:"r"`
	} `json:"qr"`
	Stats nodeclient.StatsResponse `json:"stats"`
}

func TestIntegration_GoToNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	nodeURL := startNodeAPI(t)
	app := newApp(t, nodeclient.New(nodeURL))
	token := loginToken(t, app)

	t.Run("full pipeline", func(t *testing.T) {
		resp := performTimeout(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3,4],[5,6]]}`), token, 5*time.Second)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			raw, _ := io.ReadAll(resp.Body)
			t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
		}

		var got matrixAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}

		original := matrix.Matrix{{1, 2}, {3, 4}, {5, 6}}
		if !matrix.AlmostEqual(got.Original, original, 0) {
			t.Fatalf("original = %v", got.Original)
		}
		if !matrix.AlmostEqual(got.Rotated, matrix.Matrix{{5, 3, 1}, {6, 4, 2}}, 0) {
			t.Fatalf("rotated = %v", got.Rotated)
		}

		product, err := matrix.MatMul(got.QR.Q, got.QR.R)
		if err != nil {
			t.Fatal(err)
		}
		if !matrix.AlmostEqual(product, original, 1e-8) {
			t.Fatal("Q R does not reconstruct A")
		}

		want := expectedStats(t, original)
		assertStats(t, got.Stats, want)
	})

	t.Run("validation stays 400", func(t *testing.T) {
		resp := performTimeout(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3]]}`), token, 5*time.Second)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			raw, _ := io.ReadAll(resp.Body)
			t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
		}
	})
}

func TestIntegration_NodeDownReturns502(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	nodeURL, stop := startNodeAPIProcess(t)
	stop()

	app := newApp(t, nodeclient.New(nodeURL))
	token := loginToken(t, app)
	resp := performTimeout(t, app, http.MethodPost, "/api/v1/matrix", strings.NewReader(`{"matrix":[[1,2],[3,4]]}`), token, 6*time.Second)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, raw)
	}
}

func startNodeAPI(t *testing.T) string {
	t.Helper()
	url, _ := startNodeAPIProcess(t)
	return url
}

func startNodeAPIProcess(t *testing.T) (string, func()) {
	t.Helper()

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not in PATH")
	}

	root := findRepoRoot(t)
	nodeDir := filepath.Join(root, "node-api")
	if _, err := os.Stat(filepath.Join(nodeDir, "node_modules", "express")); err != nil {
		t.Skip("node-api dependencies missing; run npm install in node-api")
	}

	port := freePort(t)
	cmd := exec.Command("node", "src/app.js")
	cmd.Dir = nodeDir
	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(port))
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start node-api: %v", err)
	}

	stopped := false
	stop := func() {
		if stopped || cmd.Process == nil {
			return
		}
		stopped = true
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
	t.Cleanup(stop)

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitHealthy(t, url, &stderr)
	return url, stop
}

func waitHealthy(t *testing.T, baseURL string, stderr *bytes.Buffer) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	client := &http.Client{Timeout: 200 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("node-api did not become healthy at %s\nstderr: %s", baseURL, stderr.String())
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		compose := filepath.Join(dir, "docker-compose.yml")
		app := filepath.Join(dir, "node-api", "src", "app.js")
		if _, err := os.Stat(compose); err == nil {
			if _, err := os.Stat(app); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (docker-compose.yml + node-api)")
		}
		dir = parent
	}
}

func performTimeout(t *testing.T, app *fiber.App, method, path string, body io.Reader, token string, timeout time.Duration) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: timeout})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func expectedStats(t *testing.T, original matrix.Matrix) nodeclient.StatsResponse {
	t.Helper()
	rotated, err := matrix.Rotate90Clockwise(original)
	if err != nil {
		t.Fatal(err)
	}
	qr, err := matrix.FactorizeQR(original)
	if err != nil {
		t.Fatal(err)
	}

	matrices := map[string]matrix.Matrix{
		"q":       qr.Q,
		"r":       qr.R,
		"rotated": rotated,
	}

	min := math.Inf(1)
	max := math.Inf(-1)
	sum := 0.0
	count := 0.0
	diag := make(map[string]bool, len(matrices))
	for name, m := range matrices {
		diag[name] = isDiagonal(m, 1e-9)
		for _, row := range m {
			for _, v := range row {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
				sum += v
				count++
			}
		}
	}
	return nodeclient.StatsResponse{
		Max: max, Min: min, Sum: sum, Average: sum / count, Diagonal: diag,
	}
}

func isDiagonal(m matrix.Matrix, eps float64) bool {
	rows, cols := matrix.Dims(m)
	if rows != cols {
		return false
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i != j && math.Abs(m[i][j]) >= eps {
				return false
			}
		}
	}
	return true
}

func assertStats(t *testing.T, got, want nodeclient.StatsResponse) {
	t.Helper()
	const eps = 1e-8
	if math.Abs(got.Max-want.Max) > eps || math.Abs(got.Min-want.Min) > eps ||
		math.Abs(got.Sum-want.Sum) > eps || math.Abs(got.Average-want.Average) > eps {
		t.Fatalf("stats = %+v, want %+v", got, want)
	}
	for _, key := range []string{"q", "r", "rotated"} {
		if got.Diagonal[key] != want.Diagonal[key] {
			t.Fatalf("diagonal[%s] = %v, want %v", key, got.Diagonal[key], want.Diagonal[key])
		}
	}
}
