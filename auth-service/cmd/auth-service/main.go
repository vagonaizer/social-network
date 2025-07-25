package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"social-network/auth-service/internal/app"
)

func main() {
	// Создаем приложение
	application := app.New()

	// Инициализируем все компоненты
	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Обработка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping application...")
		if err := application.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Запускаем приложение
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
