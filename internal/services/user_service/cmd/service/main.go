package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/app"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()

		stop()
	}()

	application := app.New(cfg)

	if err := application.Run(ctx); err != nil {
		log.Printf("auth_service stopped with error: %s", err)
		os.Exit(1)
	}
}
