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
		"Your login",
		"Your Password",
		client.WithClientID("sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"),
		client.WithLogger(logger),
		client.WithCircuitBreaker(breaker),
		client.WithDebug(false),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Получаем информацию о пользователе
	userInfo, err := client.GetMqttUserInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user info: %v", err)
	}

	// Выводим MQTT-данные
	if userInfo.MQTTUser != nil {
		fmt.Printf("MQTT Username: %s\n", userInfo.MQTTUser.Username)
		fmt.Printf("MQTT Password: %s\n", userInfo.MQTTUser.Password)
	} else {
		log.Println("MQTTUser is nil — проверьте, что /user возвращает данные")
	}

	// Получаем список зданий
	buildings, err := client.GetBuildings(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch buildings: %v", err)
	}

	// Получаем и выводим состояние всех устройств
	for _, building := range buildings {
		log.Printf("Building: %s", building.Title)
		for _, device := range building.Places {
			deviceState, err := client.GetDeviceState(context.Background(), device.ID)
			if err != nil {
				log.Printf("Failed to fetch state for device %d (%s): %v", device.ID, device.Title, err)
				continue
			}

			// Выводим состояние кондиционера
			log.Printf("Device: %s", deviceState.Title)
			log.Printf("  Temp: %.1f°C", deviceState.CurTemp)
			log.Printf("  Online: %v", deviceState.IsOnline())
			log.Printf("  IsOn: %v", deviceState.State.IsOn)
			log.Printf("  Info Text: %s", deviceState.State.Info.Text)

			// Выводим детали состояния
			for _, detail := range deviceState.State.Details {
				for _, d := range detail.Details {
					if d.Text != nil {
						log.Printf("  - State Text: %s", *d.Text)
					}
					if d.Icon != nil {
						log.Printf("  - Icon: %s", *d.Icon)
					}
				}
			}
			fmt.Println()
		}
	}
}
