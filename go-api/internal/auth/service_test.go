package auth

import (
	"errors"
	"testing"
	"time"
)

func TestLoginAndParse(t *testing.T) {
	t.Parallel()

	svc := mustService(t, time.Hour)
	token, exp, err := svc.Login("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if time.Until(exp) < 50*time.Minute {
		t.Fatalf("expiry too soon: %v", exp)
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "admin" {
		t.Fatalf("sub = %s", claims.Subject)
	}
}

func TestLoginRejectsBadPassword(t *testing.T) {
	t.Parallel()

	svc := mustService(t, time.Hour)
	if _, _, err := svc.Login("admin", "nope"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v", err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	t.Parallel()

	svc := mustService(t, time.Hour)
	if _, err := svc.Parse("not-a-jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v", err)
	}
}

func TestParseRejectsExpired(t *testing.T) {
	t.Parallel()

	svc := mustService(t, time.Millisecond)
	token, _, err := svc.Login("admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := svc.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want expired", err)
	}
}

func TestNewRequiresSecret(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Username: "a", Password: "b"})
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("error = %v", err)
	}
}

func mustService(t *testing.T, ttl time.Duration) *Service {
	t.Helper()
	svc, err := New(Config{
		Secret:   "test-secret-not-for-production",
		Username: "admin",
		Password: "secret",
		TTL:      ttl,
	})
	if err != nil {
		t.Fatal(err)
	}
	return svc
}
