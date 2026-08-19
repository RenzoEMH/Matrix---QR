package main

import (
	"log"

	"matrix-qr-apis/go-api/internal/auth"
	"matrix-qr-apis/go-api/internal/config"
	apihttp "matrix-qr-apis/go-api/internal/http"
	"matrix-qr-apis/go-api/internal/nodeclient"
)

func main() {
	cfg := config.FromEnv()
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

	log.Printf("go-api listening on :%s (NODE_API_URL=%s)", cfg.Port, cfg.NodeAPIURL)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
