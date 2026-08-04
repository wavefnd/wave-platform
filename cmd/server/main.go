package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wavefnd/wave-platform/internal/platform"
)

func main() {
	configPath := os.Getenv("WAVE_PLATFORM_CONFIG")
	if configPath == "" {
		configPath = "./config/development.xml"
	}

	application, err := platform.New(configPath)
	if err != nil {
		log.Fatalf("initialize Wave Platform: %v", err)
	}

	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("close Wave Platform: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run Wave Platform: %v", err)
	}
}
