# gKeep — Менеджер паролей (Client-Server)


Клиент и сервер взаимодействуют по защищенному каналу (TLS). Данные передаются через gRPC-вызовы и хранятся в PostgreSQL.

На клиенте реализован TUI-интерфейс при помощью пакета https://github.com/rivo/tview

---

## Требования

- **Go**: версия 1.22+ (для сборки из исходников)
- **PostgreSQL**: версия 14+ (для запуска сервера)
- **OpenSSL**: для генерации TLS-сертификатов

---

## Генерация TLS-сертификатов

Для работы клиента и сервера необходимы сертификаты. Сгенерируйте их командой:

```bash
make cert
```

Или вручную:

```bash
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/C=RU/ST=State/L=City/O=gKeep/CN=gKeep Root CA" -out ca.crt
openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/C=RU/ST=State/L=City/O=gKeep/CN=localhost" -out server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
```

---

## Сборка проекта

Используя `Makefile`:
```bash
# Сборка сервера
make build-server

# Сборка клиента
make build-client
```

Напрямую через `go build`:
```bash
# Сборка сервера
go build -ldflags "-X 'main.Version=$(git describe --tags --always)' -X 'main.BuildDate=$(date +'%Y-%m-%d')'" -o cmd/server/server cmd/server/main.go

# Сборка клиента
go build -ldflags "-X 'main.Version=$(git describe --tags --always)' -X 'main.BuildDate=$(date +'%Y-%m-%d')'" -o cmd/client/client cmd/client/main.go
```

---

## Запуск и конфигурация

Приложения настраиваются через флаги командной строки или переменные окружения. Приоритет отдается флагам.

### Конфигурация сервера

| Флаг | Переменная окружения | По умолчанию | Описание |
|---|---|---|---|
| `-addr` | `ADDRESS` | `localhost:50051` | Адрес gRPC-сервера |
| `-db` | `DATABASE_URI` | *обязательный* | Строка подключения к PostgreSQL |
| `-jwt` | `JWT_TOKEN` | *обязательный* | Секретный ключ для подписи JWT |
| `-srv-tls-cert` | `SERVER_TLS_CERT` | `server.crt` | Путь к TLS-сертификату сервера |
| `-srv-tls-key` | `SERVER_TLS_KEY` | `server.key` | Путь к TLS-ключу сервера |

#### Запуск сервера

**Linux / macOS:**
```bash
./cmd/server/server \
  -addr="localhost:50051" \
  -db="postgres://gkeep:gkeep@localhost:5432/gkeep?sslmode=disable" \
  -jwt="supersecretjwtkey" \
  -srv-tls-cert="server.crt" \
  -srv-tls-key="server.key"
```

Или через `Makefile`:
```bash
make build-server-and-listen
```

**Windows (PowerShell / CMD):**
```powershell
.\bin\server-windows-amd64.exe -addr="localhost:50051" -db="postgres://gkeep:gkeep@localhost:5432/gkeep?sslmode=disable" -jwt="supersecretjwtkey"
```

---

### Конфигурация клиента

| Флаг | Переменная окружения | По умолчанию | Описание |
|---|---|---|---|
| `-addr` | `ADDRESS` | `localhost:50051` | Адрес подключения к gRPC-серверу |
| `-client-tls-cert` | `CLIENT_TLS_CERT` | `ca.crt` | Путь к корневому CA-сертификату |

#### Запуск клиента

**Linux / macOS:**
```bash
./cmd/client/client \
  -addr="localhost:50051" \
  -client-tls-cert="ca.crt"
```

Или через `Makefile`:
```bash
make build-client-and-listen
```

**Windows (PowerShell / CMD):**
```powershell
.\bin\client-windows-amd64.exe -addr="localhost:50051" -client-tls-cert="ca.crt"
```

---

## Тестирование

Запуск тестов проекта:

```bash
go test -v ./...
```