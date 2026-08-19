package auth

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "matrix-qr-go-api"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrMissingSecret      = errors.New("jwt secret is required")
	ErrMissingUser        = errors.New("jwt username and password are required")
)

// Config holds HMAC signing material and the single demo user.
type Config struct {
	Secret   string
	Username string
	Password string
	TTL      time.Duration
}

// Service issues and validates HS256 JWTs.
type Service struct {
	secret   []byte
	username string
	password string
	ttl      time.Duration
}

// New builds a Service. Secret must be non-empty.
func New(cfg Config) (*Service, error) {
	if cfg.Secret == "" {
		return nil, ErrMissingSecret
	}
	if cfg.Username == "" || cfg.Password == "" {
		return nil, ErrMissingUser
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Service{
		secret:   []byte(cfg.Secret),
		username: cfg.Username,
		password: cfg.Password,
		ttl:      ttl,
	}, nil
}

// Login returns a signed token if username and password match.
func (s *Service) Login(username, password string) (token string, expiresAt time.Time, err error) {
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(s.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) == 1
	if userOK && passOK {
		return s.issue(username)
	}
	return "", time.Time{}, ErrInvalidCredentials
}

// Parse validates signature, expiry and issuer.
func (s *Service) Parse(token string) (*jwt.RegisteredClaims, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) {
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
	)
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) issue(subject string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	})
	signed, err := t.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}
