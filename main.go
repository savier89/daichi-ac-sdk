package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/savier89/daichi-ac-sdk/client"
)

func main() {
	logger := client.NewLogger(client.LogDebug, os.Stderr)

	breaker := client.NewCircuitBreaker(client.CircuitBreakerConfig{
		Name:        "daichi_api_breaker",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		IsError: func(err error) bool {
			return err != nil
		},
	})

	// Создаем клиент
	client, err := client.NewAuthorizedDaichiClient(
		context.Background(),
		"Your Login",
		"Your Password",
		client.WithClientID("sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"),
		client.WithLogger(logger),
		client.WithCircuitBreaker(breaker),
		client.WithDebug(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Получаем информацию о пользователе
	userInfo, err := client.GetMqttUserInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user info: %v", err)
	}

	// ✅ Выводим MQTT-данные
	if userInfo.MQTTUser != nil {
		fmt.Printf("MQTT Username: %s\n", userInfo.MQTTUser.Username)
		fmt.Printf("MQTT Password: %s\n", userInfo.MQTTUser.Password)
	} else {
		fmt.Println("MQTTUser is nil — проверьте, что /user возвращает данные в data.mqttUser")
	}

	buildings, err := client.GetBuildings(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch buildings: %v", err)
	}

	// Получаем конкретное устройство
	if len(buildings) > 0 && len(buildings[0].Places) > 0 {
		kitchenAC := buildings[0].Places[0]

		// Получаем актуальное состояние через /devices/{id}
		device, err := client.GetDeviceState(context.Background(), kitchenAC.ID)
		if err != nil {
			log.Fatalf("Failed to fetch device state: %v", err)
		}

		// Выводим текущее состояние
		fmt.Printf("Device: %s, Temp: %.1f°C, Online: %v\n",
			device.Title, device.CurTemp, device.IsOnline())

		// Выводим детали состояния
		for _, detail := range device.State.Details {
			for _, d := range detail.Details {
				if d.Text != nil {
					fmt.Printf("  - State Text: %s\n", *d.Text)
				}
				if d.Icon != nil {
					fmt.Printf("  - Icon: %s\n", *d.Icon)
				}
			}
		}
	}
}
