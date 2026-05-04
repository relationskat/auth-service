package main

import (
	"github.com/gookit/slog"
	"github.com/relationskat/auth-service/internal/config"
)

func main() {
	_, err := config.New()
	if err != nil {
		slog.Fatal("failed to load config")
	}

	slog.Info("succesfull added cfg")
}
