package http

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"matrix-qr-apis/go-api/internal/auth"
	"matrix-qr-apis/go-api/internal/matrix"
	"matrix-qr-apis/go-api/internal/nodeclient"
)

const maxMatrixDim = 50

// StatsClient is the Node stats API used after QR and rotation.
type StatsClient interface {
	ComputeStats(ctx context.Context, req nodeclient.StatsRequest) (*nodeclient.StatsResponse, error)
}

type errorBody struct {
	Error string `json:"error"`
}

type matrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type qrBody struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

type matrixResponse struct {
	Original [][]float64              `json:"original"`
	Rotated  [][]float64              `json:"rotated"`
	QR       qrBody                   `json:"qr"`
	Stats    nodeclient.StatsResponse `json:"stats"`
}

// NewApp builds the Fiber application with JSON errors and JWT on the public API.
func NewApp(node StatsClient, tokens *auth.Service, corsOrigins []string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   "matrix-qr-go-api",
		BodyLimit: 1 << 20,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
				if code < 500 {
					msg = fe.Message
				}
			}
			return c.Status(code).JSON(errorBody{Error: msg})
		},
	})
	if len(corsOrigins) > 0 {
		app.Use(cors.New(cors.Config{
			AllowOrigins: corsOrigins,
			AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
			AllowMethods: []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodOptions},
		}))
	}
	Register(app, node, tokens)
	return app
}

// Register mounts public routes on the Fiber app.
func Register(app *fiber.App, node StatsClient, tokens *auth.Service) {
	app.Get("/health", health)
	app.Post("/auth/login", login(tokens))
	app.Post("/api/v1/matrix", requireAuth(tokens), processMatrix(node))
}

func health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "go-api",
	})
}

func processMatrix(node StatsClient) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req matrixRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorBody{Error: "invalid json body"})
		}

		src := matrix.Matrix(req.Matrix)
		if err := validateIncoming(src); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorBody{Error: err.Error()})
		}

		rotated, err := matrix.Rotate90Clockwise(src)
		if err != nil {
			return internalError(c, err)
		}

		qr, err := matrix.FactorizeQR(src)
		if err != nil {
			return internalError(c, err)
		}

		stats, err := node.ComputeStats(c.Context(), nodeclient.StatsRequest{
			Matrices: map[string][][]float64{
				"q":       qr.Q,
				"r":       qr.R,
				"rotated": rotated,
			},
		})
		if err != nil {
			log.Printf("stats service: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(errorBody{Error: "stats service unavailable"})
		}

		return c.JSON(matrixResponse{
			Original: src,
			Rotated:  rotated,
			QR:       qrBody{Q: qr.Q, R: qr.R},
			Stats:    *stats,
		})
	}
}

func validateIncoming(a matrix.Matrix) error {
	if err := matrix.Validate(a); err != nil {
		return err
	}
	rows, cols := matrix.Dims(a)
	if rows > maxMatrixDim || cols > maxMatrixDim {
		return fmt.Errorf("matrix exceeds %d×%d limit", maxMatrixDim, maxMatrixDim)
	}
	return nil
}

func internalError(c fiber.Ctx, err error) error {
	log.Printf("internal error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(errorBody{Error: "internal server error"})
}
