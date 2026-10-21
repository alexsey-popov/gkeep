# Подключение к БД
DATABASE_DSN="postgres://gkeep:gkeep@localhost:5432/gkeep?sslmode=disable"

build-server:
	go build -o="cmd/server/server" cmd/server/main.go

build-server-and-listen: build-server
	./cmd/server/server -d="host=localhost user=gkeep password=gkeep dbname=gkeep sslmode=disable"

# Запускаем миграции
migrate-up:
	migrate -database $(DATABASE_DSN) -path ./migrations up

# Запускаем откат миграции
migrate-down:
	migrate -database $(DATABASE_DSN) -path ./migrations down

# Генерация кода из proto файла
proto:
	protoc \
--go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--go_opt=default_api_level=API_OPAQUE \
internal/handler/proto/user.proto