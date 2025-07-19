# 🌟 Daichi AC SDK for Go

**A powerful Go SDK for managing Daichi air conditioners via their Wi-Fi module.**  
This SDK was developed with the assistance of **Qwen3 AI** and has been tested for real-world use.  
While it may contain minor inaccuracies, all examples and `curl` commands are verified and functional.  
Perfect for developers integrating with Daichi's API for smart HVAC control.

---

## 🔧 Key Features

- ✅ **Authentication** via `/token` endpoint (POST)
- ✅ **User Profile** retrieval (`GET /user`)
- ✅ **Building & Device Management** (`GET /buildings` with nested devices)
- ✅ **Robust Error Handling** for 404, 405, 401, and 400 errors
- ✅ **Circuit Breaker** integration for fault tolerance
- ✅ **Flexible Logging** (enable/disable DEBUG mode)
- ✅ **Type-Safe Structures** for MQTT credentials, user data, and device states

---

## 📦 Installation

```bash
go get github.com/savier89/daichi-ac-sdk
```

---

## 🚀 Usage Example

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

// Create authenticated client
client, err := client.NewAuthorizedDaichiClient(
    context.Background(),
    "your-username@gmail.com",
    "your-password",
    client.WithClientID("your-client-id"),
    client.WithDebug(true),
    client.WithCircuitBreaker(breaker),
)
if err != nil {
    log.Fatal(err)
}

// Get user info
userInfo, err := client.GetMqttUserInfo(context.Background())
if err != nil {
    log.Fatalf("Failed to fetch user info: %v", err)
}

// Get buildings with devices
buildings, err := client.GetBuildings(context.Background())
if err != nil {
    log.Fatalf("Failed to fetch buildings: %v", err)
}

// Output structured data
userJSON, _ := json.MarshalIndent(userInfo, "", "  ")
fmt.Println("User Info JSON:")
fmt.Println(string(userJSON))

buildingsJSON, _ := json.MarshalIndent(buildings, "", "  ")
fmt.Println("Buildings JSON:")
fmt.Println(string(buildingsJSON))
```

---

## 📁 API Methods

### 1. **Authentication**
```go
func (c *DaichiClient) GetToken(ctx context.Context) error
```
- **Endpoint**: `POST /token`
- **Parameters**: `grant_type=password`, `email`, `password`, `clientId`
- **Example**:
  ```bash
  curl -v -X POST "https://web.daichicloud.ru/api/v4/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password" \
    -d "email=your-username" \
    -d "password=your-password" \
    -d "clientId=sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
  ```

### 2. **User Info**
```go
func (c *DaichiClient) GetUserInfo(ctx context.Context) (*DaichiUser, error)
```
- **Endpoint**: `GET /user`
- **Returns**: User ID, MQTT credentials, personal details, and access info

### 3. **Building Management**
```go
func (c *DaichiClient) GetBuildings(ctx context.Context) ([]DaichiBuilding, error)
```
- **Endpoint**: `GET /buildings`
- **Returns**: 
  - Building metadata (ID, title, coordinates)
  - Nested devices with:
    - Current temperature (`curTemp`)
    - Device state (`state.ison`, `state.info.text`)
    - Control features (`features.canChangeWiFiFromServer`, `features.serverTimerSupported`)

---

## ⚠️ Error Handling

- `ErrMissingCredentials` — Missing `username/password`
- `ErrTokenNotFound` — Token not in response
- `ErrMethodNotAllowed` — 405 errors (e.g., using `GET` on `/token`)
- `ErrEndpointNotFound` — 404 errors (verify API paths)
- `ErrTokenExpired` — 401 Unauthorized (automatic token refresh support)
- `ErrCircuitBreakerOpen` — Circuit Breaker triggered (protects against cascading failures)

---

## 🛠 Advanced Features

### 🧠 AI-Generated with Qwen3
- Structs automatically mapped from TypeScript Zod schemas
- JSON responses validated against real API data
- Circuit Breaker patterns for production resilience

### 📊 Real-Time Device State
```go
type DaichiBuildingDevice struct {
    ID                int             `json:"id"`
    CurTemp           float64         `json:"curTemp"`         // Current temperature
    State             DeviceState     `json:"state"`            // Power status + icons
    Features          map[string]bool `json:"features"`         // Control capabilities
}
```

### 🔐 Secure Token Management
- Token auto-refresh on 401 Unauthorized
- Bearer token injection in headers
- JWT lifetime tracking

---

## 🧪 Debugging Recommendations

1. **Token Issues**:
   ```bash
   curl -v -X POST "https://web.daichicloud.ru/api/v4/token" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password&email=your@email.com&password=your-pass&clientId=sOJO7B6..."
   ```

2. **Building Data**:
   ```bash
   curl -v -X GET "https://web.daichicloud.ru/api/v4/buildings" \
     -H "Authorization: Bearer <your-token>" \
     -H "Accept: application/json"
   ```

3. **Common Fixes**:
   - For 404 errors: Try `/api/v4/buildings` or `/mqtt/buildings`
   - For 405 errors: Ensure `POST` is used for `/token`
   - For nil MQTTUser: Verify server returns `mqttUser` in `/user` response

---

## 📈 Why Use This SDK?

- **Type-Safe**: All structs validated against real API responses
- **Production-Ready**: Circuit Breaker + Retries + Logging
- **Developer-Friendly**: Clean interfaces, functional options, and error types
- **Well-Documented**: Full examples and debugging guides

---

## 📄 License

MIT License — [View License](LICENSE)

---

## 🤝 Want a Production-Ready Solution?

This SDK was built with AI assistance but **works** with real Daichi API endpoints.  
If you need:
- **Full API integration**
- **Custom device control logic**
- **Mobile app backend**
- **Webhook integrations**

📩 **Contact me** for a ready-to-use enterprise solution  
💼 **Buy the full project under key** for guaranteed stability and support  
💻 Get a fully tested, production-grade API client

---

## 🛡️ Disclaimer

> This SDK was generated with **Qwen3 AI**.  
> While all API methods and examples are tested and functional, some JSON struct fields may need adjustment based on your specific use case.  
> **Perfect for developers** who want to integrate quickly with a working foundation.

---

## 📈 Ready to Build Smart HVAC Systems?

**Start today with this foundation**  
**Or go production-ready with a custom solution**

Let’s bring your AC systems online — smart, secure, and scalable.

---

### 📌 Pro Tip
Use `WithDebug(true)` to see:
```text
[INFO] Token request URL: https://web.daichicloud.ru/api/v4/token
[INFO] User info received: {"done":true,"data":{"id":120980,"token":"...","mqttUser":{"username":"...","password":"..."}}
```

---

**Made with ❤️ and Qwen3 AI**  
**For the latest updates, visit**: [GitHub Project](https://github.com/savier89/daichi-ac-sdk)  
**Need enterprise support?** [Contact Me](a.vedeneev89@gmail.com)