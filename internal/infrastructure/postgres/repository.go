package database

import (
	"log/slog"

	model "order/internal/entity"
	"order/internal/infrastructure/cache"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository struct for database and cache
type Repository struct {
	db    *gorm.DB
	cache *cache.OrderCache
}

// NewRepository creates a new instance of Repository with database and cache dependencies
func NewRepository(db *gorm.DB, cache *cache.OrderCache) *Repository {
	return &Repository{db: db, cache: cache}
}

// getOrder retrieves an order from the database by UID including related Delivery, Payment, and Items
func (r *Repository) getOrder(uid string) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", uid).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			slog.Warn("order not found", slog.String("uid", uid))
			return nil, nil
		}
		slog.Error("db error on get order", slog.String("uid", uid), slog.String("error", err.Error()))
		return nil, err
	}

	return &order, nil
}

// GetOrderWithCache retrieves an order using cache first, falling back to database if cache miss occurs
func (r *Repository) GetOrderWithCache(uid string) (*model.Order, error) {
	if cached, ok := r.cache.Get(uid); ok {
		slog.Debug("cache hit", slog.String("uid", uid))
		return cached, nil
	}

	slog.Debug("cache miss", slog.String("uid", uid))
	order, err := r.getOrder(uid)
	if order != nil {
		r.cache.Set(uid, order)
	}
	return order, err
}

// insertOrder inserts a new order and its related entities (Delivery, Payment, Items) into the database
func (r *Repository) insertOrder(tx *gorm.DB, order *model.Order) error {
	if err := tx.Create(order).Error; err != nil {
		return err
	}

	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Delivery).Error; err != nil {
		return err
	}

	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Payment).Error; err != nil {
		return err
	}

	for i := range order.Items {
		item := &order.Items[i]
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_uid"}, {Name: "rid"}},
			UpdateAll: true,
		}).Create(item).Error; err != nil {
			return err
		}
	}

	return nil
}

// updateOrder updates an existing order and its related entities (Delivery, Payment, Items) in the database
func (r *Repository) updateOrder(tx *gorm.DB, order *model.Order) error {
	if err := tx.Save(order).Error; err != nil {
		return err
	}

	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Delivery).Error; err != nil {
		return err
	}

	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Payment).Error; err != nil {
		return err
	}

	for i := range order.Items {
		item := &order.Items[i]
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_uid"}, {Name: "rid"}},
			UpdateAll: true,
		}).Create(item).Error; err != nil {
			return err
		}
	}

	return nil
}

// AddOrder inserts a new order or updates an existing one within a database transaction
// It also handles cache invalidation and refreshes the cache after the operation
func (r *Repository) AddOrder(order *model.Order) (*model.Order, error) {
	existingOrder, err := r.getOrder(order.OrderUID)
	if err != nil {
		return nil, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if existingOrder == nil {
			return r.insertOrder(tx, order)
		}
		return r.updateOrder(tx, order)

	})

	if err != nil {
		slog.Error("failed to add order", slog.String("uid", order.OrderUID), slog.String("error", err.Error()))
		return nil, err
	}

	// Invalidate and refresh cache
	r.cache.Delete(order.OrderUID)
	updatedOrder, err := r.getOrder(order.OrderUID)
	if err != nil {
		return nil, err
	}
	r.cache.Set(order.OrderUID, updatedOrder)

	return updatedOrder, nil
}
