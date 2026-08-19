package http

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"matrix-qr-apis/go-api/internal/auth"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

func login(tokens *auth.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req loginRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorBody{Error: "invalid json body"})
		}

		token, exp, err := tokens.Login(req.Username, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				return c.Status(fiber.StatusUnauthorized).JSON(errorBody{Error: "invalid credentials"})
			}
			return internalError(c, err)
		}

		expiresIn := int64(time.Until(exp).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
		return c.JSON(loginResponse{
			Token:     token,
			ExpiresIn: expiresIn,
		})
	}
}

func requireAuth(tokens *auth.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		scheme, raw, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || raw == "" {
			return unauthorized(c)
		}
		if _, err := tokens.Parse(raw); err != nil {
			return unauthorized(c)
		}
		return c.Next()
	}
}

func unauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(errorBody{Error: "unauthorized"})
}
