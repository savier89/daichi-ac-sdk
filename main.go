package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/savier89/daichi-ac-sdk/client"
)

func main() {
	debugEnabled := true
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

	client, err := client.NewAuthorizedDaichiClient(
		context.Background(),
		"Your Login from web.daichicloud.ru",
		"Your password",
		// Default ClientID is sOJO7B6SqgaKudTfCzqLAy540cCuDzpI
		client.WithClientID("sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"),
		loggerOption,
		client.WithCircuitBreaker(breaker),
	)
	if err != nil {
		log.Fatalf("Failed to create authorized client: %v", err)
	}

	// Получаем информацию о пользователе
	userInfo, err := client.GetMqttUserInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user info: %v", err)
	}

	// Выводим JSON
	userJSON, _ := json.MarshalIndent(userInfo, "", "  ")
	fmt.Println("User Info JSON:")
	fmt.Println(string(userJSON))

	// ✅ Пример доступа к MQTTUser
	if userInfo.MQTTUser != nil {
		fmt.Println("MQTT Username:", userInfo.MQTTUser.Username)
		fmt.Println("MQTT Password:", userInfo.MQTTUser.Password)
	} else {
		log.Println("MQTTUser is nil")
	}

	// Получаем список зданий
	buildings, err := client.GetBuildings(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch buildings: %v", err)
	}

	// Выводим JSON
	buildingsJSON, _ := json.MarshalIndent(buildings, "", "  ")
	fmt.Println("Buildings JSON:")
	fmt.Println(string(buildingsJSON))

	// ✅ Пример доступа к данным
	for _, b := range buildings {
		fmt.Printf("Building: %s, Places: %d\n", b.Title, b.PlacesCount)
		for _, p := range b.Places {
			fmt.Printf("  Device: %s, Status: %s, Temp: %.1f°C\n", p.Title, p.Status, p.CurTemp)
		}
	}
}
