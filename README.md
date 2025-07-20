## 🇷🇺 Русский

### **daichi-ac-sdk**  
SDK для взаимодействия с кондиционерами Daichi через API.  
Реализованы методы:
- Авторизация (`/token`)
- Получение информации о пользователе (`/user`)
- Получение списка зданий (`/buildings`)
- Получение состояния устройства (`/device/{id}`)

---

### 🔧 Установка
```bash
go get github.com/savier89/daichi-ac-sdk
```

---

### 🧪 Использование
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/savier89/daichi-ac-sdk/client"
)

func main() {
	// Создаем клиент
	client, err := client.NewAuthorizedDaichiClient(
		context.Background(),
		"your-email@gmail.com",
		"your-password",
		client.WithClientID("your-client-id"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Получаем информацию о пользователе
	userInfo, err := client.GetMqttUserInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user info: %v", err)
	}
	log.Printf("MQTT Username: %s", userInfo.MQTTUser.Username)
	log.Printf("MQTT Password: %s", userInfo.MQTTUser.Password)

	// Получаем список зданий
	buildings, err := client.GetBuildings(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch buildings: %v", err)
	}

	// Получаем состояние всех устройств
	for _, b := range buildings {
		log.Printf("Building: %s", b.Title)
		for _, device := range b.Places {
			deviceState, err := client.GetDeviceState(context.Background(), device.ID)
			if err != nil {
				log.Printf("Failed to fetch state for %s: %v", device.Title, err)
				continue
			}
			log.Printf("Device: %s", deviceState.Title)
			log.Printf("  Temp: %.1f°C", deviceState.CurTemp)
			log.Printf("  Online: %v", deviceState.IsOnline())
			log.Printf("  IsOn: %v", deviceState.State.IsOn)
			log.Printf("  Info Text: %s", deviceState.State.Info.Text)
		}
	}
}
```

---

### 📁 Структура проекта
```
daichi-ac-sdk/
├── client/
│   ├── auth_roundtripper.go
│   ├── circuit_breaker.go
│   ├── device.go
│   ├── errors.go
│   ├── http_client.go
│   ├── logger.go
│   └── authorized_client.go
├── main.go
└── README.md
```

---

### 🌐 API Методы
| Метод | Описание |
|-------|----------|
| `GetToken` | Авторизация через `/token` |
| `GetUserInfo` | Получение данных пользователя через `/user` |
| `GetBuildings` | Получение списка зданий через `/buildings` |
| `GetDeviceState` | Получение состояния устройства через `/device/{id}` |

---

### 📋 Логирование
- Поддерживает уровни: `LogNone`, `LogError`, `LogWarn`, `LogInfo`, `LogDebug`
- Цвета:  
  - `DEBUG` — Cyan  
  - `INFO` — Green  
  - `WARN` — Yellow  
  - `ERROR` — Red

---

### 🛡️ Обработка ошибок
| Ошибка | Описание |
|--------|----------|
| `ErrMissingCredentials` | Логин и пароль не указаны |
| `ErrTokenNotFound` | Токен не найден в ответе |
| `ErrTokenRefreshFailed` | Обновление токена не выполнено |
| `ErrMethodNotAllowed` | Метод не поддерживается |
| `ErrEndpointNotFound` | URL не существует |
| `ErrInvalidAPIResponse` | Ответ API не соответствует ожидаемому формату |

---

### ⚙️ Настройка клиента
```go
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
	"your-email",
	"your-password",
	client.WithClientID("sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"),
	client.WithLogger(client.NewLogger(client.LogDebug, os.Stderr)),
	client.WithCircuitBreaker(breaker),
)
```

---

### 📡 Тестирование через `curl`
```bash
# Авторизация
curl -X POST "https://web.daichicloud.ru/api/v4/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "email=your-email" \
  -d "password=your-password" \
  -d "clientId=your-client-id"

# Получение информации о пользователе
curl -X GET "https://web.daichicloud.ru/api/v4/user" \
  -H "Authorization: Bearer <your-token>"

# Получение зданий
curl -X GET "https://web.daichicloud.ru/api/v4/buildings" \
  -H "Authorization: Bearer <your-token>"

# Получение состояния устройства
curl -X GET "https://web.daichicloud.ru/api/v4/device/203488" \
  -H "Authorization: Bearer <your-token>"
```

---

## EN English

### **daichi-ac-sdk**  
SDK for interacting with Daichi air conditioners via API.  
Implemented methods:
- Authentication (`/token`)
- User info (`/user`)
- Building list (`/buildings`)
- Device state (`/device/{id}`)

---

### 🔧 Installation
```bash
go get github.com/savier89/daichi-ac-sdk
```

---

### 🧪 Usage
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/savier89/daichi-ac-sdk/client"
)

func main() {
	// Create client
	client, err := client.NewAuthorizedDaichiClient(
		context.Background(),
		"your-email@gmail.com",
		"your-password",
		client.WithClientID("your-client-id"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Fetch user info
	userInfo, err := client.GetMqttUserInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user info: %v", err)
	}
	log.Printf("MQTT Username: %s", userInfo.MQTTUser.Username)
	log.Printf("MQTT Password: %s", userInfo.MQTTUser.Password)

	// Fetch buildings
	buildings, err := client.GetBuildings(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch buildings: %v", err)
	}

	// Fetch device states
	for _, b := range buildings {
		log.Printf("Building: %s", b.Title)
		for _, device := range b.Places {
			deviceState, err := client.GetDeviceState(context.Background(), device.ID)
			if err != nil {
				log.Printf("Failed to fetch state for %s: %v", device.Title, err)
				continue
			}
			log.Printf("Device: %s", deviceState.Title)
			log.Printf("  Temp: %.1f°C", deviceState.CurTemp)
			log.Printf("  Online: %v", deviceState.IsOnline())
			log.Printf("  IsOn: %v", deviceState.State.IsOn)
			log.Printf("  Info Text: %s", deviceState.State.Info.Text)
		}
	}
}
```

---

### 📁 Project Structure
```
daichi-ac-sdk/
├── client/
│   ├── auth_roundtripper.go
│   ├── circuit_breaker.go
│   ├── device.go
│   ├── errors.go
│   ├── http_client.go
│   ├── logger.go
│   └── authorized_client.go
├── main.go
└── README.md
```

---

### 🌐 API Methods
| Method | Description |
|--------|-------------|
| `GetToken` | Authenticate via `/token` |
| `GetUserInfo` | Fetch user info via `/user` |
| `GetBuildings` | Fetch building list via `/buildings` |
| `GetDeviceState` | Fetch device state via `/device/{id}` |

---

### 📋 Logging
- Supports: `LogNone`, `LogError`, `LogWarn`, `LogInfo`, `LogDebug`
- Colors:
  - `DEBUG` — Cyan
  - `INFO` — Green
  - `WARN` — Yellow
  - `ERROR` — Red

---

### 🛡️ Error Handling
| Error | Description |
|-------|-------------|
| `ErrMissingCredentials` | Email and password not set |
| `ErrTokenNotFound` | Token not found in response |
| `ErrTokenRefreshFailed` | Token refresh failed |
| `ErrMethodNotAllowed` | Method not supported |
| `ErrEndpointNotFound` | API endpoint not found |
| `ErrInvalidAPIResponse` | Invalid API response format |

---

### ⚙️ Client Configuration
```go
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
	"your-email",
	"your-password",
	client.WithClientID("sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"),
	client.WithLogger(client.NewLogger(client.LogDebug, os.Stderr)),
	client.WithCircuitBreaker(breaker),
)
```

---

### 📡 Testing with `curl`
```bash
# Authentication
curl -X POST "https://web.daichicloud.ru/api/v4/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "email=your-email" \
  -d "password=your-password" \
  -d "clientId=your-client-id"

# Fetch user info
curl -X GET "https://web.daichicloud.ru/api/v4/user" \
  -H "Authorization: Bearer <your-token>"

# Fetch buildings
curl -X GET "https://web.daichicloud.ru/api/v4/buildings" \
  -H "Authorization: Bearer <your-token>"

# Fetch device state
curl -X GET "https://web.daichicloud.ru/api/v4/device/203488" \
  -H "Authorization: Bearer <your-token>"
```

---
