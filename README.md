```markdown
# Daichi AC SDK for Go / Daichi AC SDK для Go
 
Go SDK for managing Daichi air conditioners via Wi-Fi module.

## Установка / Installation

```bash
go get github.com/savier89/daichi-ac-sdk
```

---

### Важно! / Important!
**Код написан нейросетью Qwen3 в рамках тестового задания.**  
**The code was written by Qwen3 neural network as part of a test task.**

---

## ✅ Основные улучшения / Key Improvements

### 1. Методы запросов / Request Methods
- **POST вместо GET для `/token`**  
  Сервер не принимает GET-запросы, используется POST-метод  
  *POST instead of GET for `/token`*  
  Server doesn't accept GET requests, POST method is used

### 2. Логирование / Logging
- Включение через `WithDebug(true)`  
  Отключение через `WithDebug(false)`  
  *Enable with `WithDebug(true)`*  
  *Disable with `WithDebug(false)`*

### 3. Валидация URL / URL Validation
- Используется `url.JoinPath` и `url.Parse`  
  *Uses `url.JoinPath` and `url.Parse`*

### 4. Обработка ошибок / Error Handling
| Код ошибки / Error Code | Сообщение / Message | Константа / Constant |
|-------------------------|---------------------|----------------------|
| 405 Method Not Allowed  | Метод не разрешен   | `ErrMethodNotAllowed`|
| 404 Not Found           | Эндпоинт не найден  | `ErrEndpointNotFound`|

### 5. Отладочная информация / Debug Info
- Вывод точного URL в логах:  
  ```go
  c.Logger("[INFO] Token request URL: %s", reqURL)
  ```

### 6. Автоматическое обновление токена / Token Auto-Refresh
- Реализовано в `auth_roundtripper.go` при получении 401 Unauthorized  
  *Implemented in `auth_roundtripper.go` when receiving 401 Unauthorized*

---

## 🔧 Рекомендации по отладке / Debugging Tips

### 1. Проверка токена через curl / Token Check with curl
```bash
curl -v -X POST "https://web.daichicloud.ru/api/v4/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "email=your-username" \
  -d "password=your-password" \
  -d "clientId=sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
```

---

## ⚠️ Обратите внимание / Note
При получении ошибки `405 Method Not Allowed`:
- Убедитесь, что используется POST-метод
- Проверьте Content-Type заголовок
- Убедитесь в правильности параметров запроса

*When receiving `405 Method Not Allowed` error:*
- Make sure POST method is used
- Check Content-Type header
- Verify request parameters

---

```