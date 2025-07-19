# Daichi AC SDK for Go

Go SDK для управления кондиционерами Daichi через Wi-Fi модуль.

## Установка

```bash
go get github.com/savier89/daichi-ac-sdk



---

### 🧨 **Что исправлено и улучшено**
1. **POST-запросы для `/token`**:
   - Сервер не принимает `GET`, поэтому используется `POST`.

2. **DEBUG-логирование**:
   - Включается через `WithDebug(true)`.
   - Отключается через `WithDebug(false)`.

3. **Валидация URL**:
   - Используется `url.JoinPath` и проверка через `url.Parse`.

4. **Обработка ошибок**:
   - `405 Method Not Allowed` → `ErrMethodNotAllowed`
   - `404 Not Found` → `ErrEndpointNotFound`

5. **Логирование точного URL**:
   - Везде выводится итоговый URL для отладки:
     ```go
     c.Logger("[INFO] Token request URL: %s", reqURL)
     ```

6. **Автоматическое обновление токена при 401 Unauthorized**:
   - Реализовано в `auth_roundtripper.go`.

---

### 📌 **Рекомендации по отладке**
1. **Проверьте токен через `curl`**:
   ```bash
   curl -v -X POST "https://web.daichicloud.ru/api/v4/token " \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password" \
     -d "email=your-username" \
     -d "password=your-password" \
     -d "clientId=sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
