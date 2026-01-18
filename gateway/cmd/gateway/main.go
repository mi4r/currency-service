package main

import (
	"log"

	"github.com/mi4r/currency-service/gateway/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
