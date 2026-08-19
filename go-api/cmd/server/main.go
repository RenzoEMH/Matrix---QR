package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matrix-qr-apis/go-api/internal/auth"
	"matrix-qr-apis/go-api/internal/config"
	apihttp "matrix-qr-apis/go-api/internal/http"
	"matrix-qr-apis/go-api/internal/nodeclient"
)

func main() {
	cfg := config.FromEnv()
	if cfg.JWTSecret == "dev-only-change-me-please" {
		log.Println("warning: using default JWT_SECRET; set a strong secret in production")
	}

	tokens, err := auth.New(auth.Config{
		Secret:   cfg.JWTSecret,
		Username: cfg.JWTUsername,
		Password: cfg.JWTPassword,
		TTL:      cfg.JWTTTL,
	})
	if err != nil {
		log.Fatal(err)
	}

	app := apihttp.NewApp(nodeclient.New(cfg.NodeAPIURL), tokens, cfg.CORSOrigins)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("go-api listening on :%s (NODE_API_URL=%s)", cfg.Port, cfg.NodeAPIURL)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
