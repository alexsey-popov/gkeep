// migrations Пакет для выполнения миграций. Файлы миграций зашиваются в бинарник
package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed *.sql
var FS embed.FS

// Up Выполнение миграций БД
func Up(db *sqlx.DB) error {
	// 1. Создаем источник на базе вложенных файлов
	d, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("ошибка при создании источника файлов миграции: %w", err)
	}

	// Создаём драйвер для миграций
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка при создании драйвера БД: %w", err)
	}

	//   Создаём объект миграции на основе файлов с миграциями и подключения
	m, err := migrate.NewWithInstance("iofs", d, "pgx", driver)
	if err != nil {
		return fmt.Errorf("ошибка при подготовке к миграций БД: %w", err)
	}

	// Проводим миграции
	err = m.Up()
	// Ошибку migrate.ErrNoChange пропускаем
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("ошибка при запуске миграций БД: %w", err)
	}

	return nil
}
