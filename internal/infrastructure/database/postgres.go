package database

import (
	"fmt"
	"order/internal/config"
	"order/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresDB initializes a new PostgreSQL connection, performs migrations,
// and ensures necessary indexes exist.
func NewPostgresDB(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB object: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	if err := db.AutoMigrate(
		&model.Order{},
		&model.Delivery{},
		&model.Payment{},
		&model.Item{},
	); err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %w", err)
	}

	return db, nil
}
