package http_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"matrix-qr-apis/go-api/internal/auth"
	apihttp "matrix-qr-apis/go-api/internal/http"
)

const (
	testUser = "admin"
	testPass = "secret"
)

func testAuth(t *testing.T) *auth.Service {
	t.Helper()
	svc, err := auth.New(auth.Config{
		Secret:   "test-secret-not-for-production",
		Username: testUser,
		Password: testPass,
		TTL:      time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func newApp(t *testing.T, node apihttp.StatsClient) *fiber.App {
	t.Helper()
	return apihttp.NewApp(node, testAuth(t), nil)
}

func loginToken(t *testing.T, app *fiber.App) string {
	t.Helper()
	body := `{"username":"` + testUser + `","password":"` + testPass + `"}`
	resp := perform(t, app, http.MethodPost, "/auth/login", strings.NewReader(body), "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("login status = %d, body = %s", resp.StatusCode, raw)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Token == "" {
		t.Fatal("login returned empty token")
	}
	return out.Token
}

func perform(t *testing.T, app *fiber.App, method, path string, body io.Reader, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
