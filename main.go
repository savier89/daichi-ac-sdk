package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/savier89/daichi-ac-sdk/client"
)

func main() {
	// ✅ Включаем или выключаем DEBUG-логи
	debugEnabled := true // поменяйте на false, чтобы выключить логи
	loggerOption := client.WithDebug(debugEnabled)

	breaker := client.NewCircuitBreaker(client.CircuitBreakerConfig{
		Name:        "daichi_api_breaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		IsError: func(err error) bool {
			return err != nil
		},
	})

	// Создаем клиент с опцией включения DEBUG
	client, err := client.NewAuthorizedDaichiClient(
		context.Background(),
		"your-username@gmail.com",
		"your-password",
		client.WithClientID("your-client-id"),
		loggerOption,
		client.WithCircuitBreaker(breaker),
	)
	if err != nil {
		log.Fatalf("Failed to create authorized client: %v", err)
	}

	// Проверяем доступность API
	response, err := client.Ping(context.Background())
	if err != nil {
		log.Fatalf("API ping failed: %v", err)
	}

	fmt.Println("API Response:", response)
}
