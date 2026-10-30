package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexsey-popov/gkeep/internal/client"
	"github.com/alexsey-popov/gkeep/internal/config"
	pb "github.com/alexsey-popov/gkeep/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	Version   = "0.0.0"
	BuildDate = "0000-00-00"
)

func main() {
	cfg, err := config.NewClient(os.Args[1:], os.LookupEnv)
	if err != nil {
		log.Fatalf("ошибка инициализации конфигурации клиента: %v", err)
	}

	// Получает TLS сертификат
	tlsCreds, err := credentials.NewClientTLSFromFile(cfg.TLSCert, "")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(tlsCreds),
	)
	if err != nil {
		log.Fatalf("не удалось подключиться к серверу gRPC по адресу %s: %v", cfg.Address, err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("ошибка закрытия соединения: %v\n", closeErr)
		}
	}()

	usersClient := pb.NewAPIClient(conn)
	app := client.NewApp(usersClient, fmt.Sprintf("%s %s", Version, BuildDate))

	if err := app.Run(); err != nil {
		log.Fatalf("ошибка выполнения приложения: %v", err)
	}
}
