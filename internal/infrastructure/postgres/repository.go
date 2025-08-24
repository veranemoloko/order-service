package database

import (
	"log/slog"

	model "order/internal/entity"
	"order/internal/infrastructure/cache"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository provides methods to work with orders in the database and cache
type Repository struct {
	db    *gorm.DB
	cache *cache.OrderCache
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB, cache *cache.OrderCache) *Repository {
	return &Repository{db: db, cache: cache}
}

// GetOrder retrieves an order directly from the database along with related entities
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

// GetOrderWithCache first checks the cache, then fetches from the database if missing
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

// InsertOrUpdateOrder inserts or updates an order and all its related entities
func (r *Repository) InsertOrUpdateOrder(order *model.Order) (*model.Order, error) {

	err := r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(order).Error; err != nil {
			return err
		}

		order.Delivery.OrderUID = order.OrderUID
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Delivery).Error; err != nil {
			return err
		}

		order.Payment.OrderUID = order.OrderUID
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&order.Payment).Error; err != nil {
			return err
		}

		for i := range order.Items {
			item := &order.Items[i]
			item.OrderUID = order.OrderUID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "order_uid"}, {Name: "rid"}},
				UpdateAll: true,
			}).Create(item).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to insert/update order", slog.String("uid", order.OrderUID), slog.String("error", err.Error()))
		return nil, err
	}

	// Update cache after successful transaction (remove and reload)
	r.cache.Delete(order.OrderUID)
	updatedOrder, err := r.getOrder(order.OrderUID)
	if err != nil {
		return nil, err
	}
	r.cache.Set(order.OrderUID, updatedOrder)
	return updatedOrder, nil
}
