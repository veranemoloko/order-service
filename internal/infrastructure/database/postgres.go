package database

import (
	"fmt"
	"order/internal/config"
	model "order/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresDB initializes a new PostgreSQL connection, performs migrations,
// and ensures necessary indexes exist.
func NewPostgresDB(cfg config.DBConfig) (*gorm.DB, error) {
	// Build the DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port,
	)

	// Open the database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	// Get the underlying sql.DB object for pinging
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB object: %w", err)
	}

	// Ping the database to verify the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&model.Order{},
		&model.Delivery{},
		&model.Payment{},
		&model.Item{},
	); err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %w", err)
	}

	// Create a unique index for (order_uid, rid) in the items table
	if err := db.Exec(`
    CREATE UNIQUE INDEX IF NOT EXISTS idx_order_rid
    ON items (order_uid, rid)
`).Error; err != nil {
		return nil, fmt.Errorf("failed to create unique index on items(order_uid, rid): %w", err)
	}

	return db, nil
}
