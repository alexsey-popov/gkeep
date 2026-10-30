SHELL := /bin/bash

# Подключение к БД
DATABASE_DSN="postgres://gkeep:gkeep@localhost:5432/gkeep?sslmode=disable"

# Секретный JWT токен
JWT="secret"

# Название файла отчётом по покрытию кода тестами
COVER_FILE?=cover.out

build-server:
	go build -o="cmd/server/server" -ldflags "-X 'main.version=${version}' -X 'main.buildDate=${date}'" cmd/server/main.go

build-server-and-listen: build-server
	./cmd/server/server -db="host=localhost user=gkeep password=gkeep dbname=gkeep sslmode=disable" -jwt="${JWT}"

build-client:
	go build -o="cmd/client/client" -ldflags "-X 'main.Version=$$(git describe --tags --always)' -X 'main.BuildDate=$$(date +'%d-%m-%Y')'" cmd/client/main.go
build-client-and-listen: build-client
	./cmd/client/client

create-migration:
	migrate create -ext sql -dir ./migrations -seq $(name)

# Запускаем миграции
migrate-up:
	migrate -database $(DATABASE_DSN) -path ./migrations up

# Запускаем откат миграции
migrate-down:
	migrate -database $(DATABASE_DSN) -path ./migrations down

# Генерация кода из proto файла
gen-proto:
	protoc \
--go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--go_opt=default_api_level=API_OPAQUE \
proto/user/user.proto proto/password/password.proto proto/schema.proto

# Создание TLS сертификата для клиента и сервера
cert:
	openssl genrsa -out ca.key 2048
	openssl req -new -x509 -days 365 -key ca.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" -out ca.crt
	openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=localhost" -out server.csr
	openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Запускаем все тесты
tests:
	go test ./... -count=1

# Подсчёт покрытия кода тестами с выводом в файл
tests-cover:
	go test ./... --coverprofile=$(COVER_FILE)

# Подсчёт покрытия кода тестами с просмотром в браузере
tests-cover-view: tests-cover
	go tool cover -html=$(COVER_FILE)

tests-cover-total: tests-cover
	go tool cover -func=$(COVER_FILE)