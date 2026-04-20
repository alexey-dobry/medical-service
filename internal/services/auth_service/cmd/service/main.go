package main

import (
	"context"
	"flag"

	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/app"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/config"
)

func main() {
	flag.Parse()

	cfg := config.MustLoad()

	application := app.New(cfg)

	application.Run(context.Background())
}
