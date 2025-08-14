package main

import (
	"fmt"
	"log"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/auth"
	"github.com/AnjuRKrishnan/fleet-tracker/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	jwtAuth := auth.NewJWTAuth(cfg.JWTSecret)
	token, err := jwtAuth.GenerateToken("user123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)
}
