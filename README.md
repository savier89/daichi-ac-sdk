# Daichi AC SDK for Go

Go SDK for managing Daichi Wi-Fi air conditioners.

## Установка

```bash
go get github.com/savier89/daichi-ac-sdk
```

## Пример использования

```go
import "github.com/savier89/daichi-ac-sdk"

client := daichi.NewDaichiClient("https://api.daichi.local ", "your-auth-key")

err := client.SetTemperature(context.Background(), "ac123", 24.0)
if err != nil {
    log.Fatal(err)
}
```
