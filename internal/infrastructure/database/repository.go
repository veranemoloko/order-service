package database

import (
	"fmt"
	model "order/internal/entity"
	"order/internal/infrastructure/cache"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository provides methods to interact with orders in the database and cache.
type Repository struct {
	db    *gorm.DB
	cache *cache.OrderCache
}

// NewRepository creates a new Repository instance.
func NewRepository(db *gorm.DB, cache *cache.OrderCache) *Repository {
	return &Repository{db: db, cache: cache}
}

// GetOrder retrieves an order by UID from the database without using cache.
func (r *Repository) GetOrder(uid string) (*model.Order, error) {
	var order model.Order

	if err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

// GetOrderWithCache retrieves an order by UID, using cache if available.
func (r *Repository) GetOrderWithCache(uid string) (*model.Order, error) {
	if order, ok := r.cache.Get(uid); ok {
		return order, nil
	}

	order, err := r.GetOrder(uid)
	if err != nil || order == nil {
		return order, err
	}

	r.cache.Set(uid, order)
	return order, nil
}

// InsertOrUpdateOrder safely inserts or updates an order in the database.
// Cache is invalidated only if there were real changes.
func (r *Repository) InsertOrUpdateOrder(order *model.Order) (*model.Order, error) {
	// 1. Check if the order already exists
	existing, err := r.GetOrder(order.OrderUID)
	if err != nil {
		return nil, fmt.Errorf("get existing order: %w", err)
	}

	// 2. If order exists and is identical → return existing, no cache update needed
	if existing != nil && reflect.DeepEqual(existing, order) {
		return existing, nil
	}

	// 3. Apply changes within a transaction
	err = r.db.Transaction(func(tx *gorm.DB) error {
		// Order
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(order).Error; err != nil {
			return fmt.Errorf("insert/update order: %w", err)
		}

		// Delivery
		order.Delivery.OrderUID = order.OrderUID
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&order.Delivery).Error; err != nil {
			return fmt.Errorf("insert/update delivery: %w", err)
		}

		// Payment
		order.Payment.OrderUID = order.OrderUID
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&order.Payment).Error; err != nil {
			return fmt.Errorf("insert/update payment: %w", err)
		}

		// Items
		for i := range order.Items {
			item := &order.Items[i]
			item.OrderUID = order.OrderUID

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "order_uid"}, {Name: "rid"}}, // unique key
				UpdateAll: true,
			}).Create(item).Error; err != nil {
				return fmt.Errorf("insert/update item: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 4. Invalidate cache only if there were changes
	if existing == nil || !reflect.DeepEqual(existing, order) {
		r.cache.Delete(order.OrderUID)
	}

	// 5. Retrieve the order again to return the latest data
	updatedOrder, err := r.GetOrder(order.OrderUID)
	if err != nil {
		return nil, fmt.Errorf("get order after insert/update: %w", err)
	}

	// 6. Update cache with the latest data
	r.cache.Set(order.OrderUID, updatedOrder)

	return updatedOrder, nil
}
