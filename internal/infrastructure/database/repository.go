package database

import (
	"log/slog"

	"order/internal/infrastructure/cache"
	"order/internal/model"

	"gorm.io/gorm"
)

// Repository wraps database and cache access for orders.
type Repository struct {
	db    *gorm.DB
	cache *cache.OrderCache
}

// NewRepository creates a new instance of Repository with database and cache dependencies.
func NewRepository(db *gorm.DB, cache *cache.OrderCache) *Repository {
	return &Repository{db: db, cache: cache}
}

// getOrder retrieves an order directly from the database by UID
// including related entities (Delivery, Payment, and Items).
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
		slog.Error("db error on get order", slog.String("uid", uid), slog.Any("err", err))
		return nil, err
	}

	slog.Debug("order loaded from db", slog.String("uid", uid))
	return &order, nil
}

// GetOrderWithCache tries to get an order from cache first.
// If cache miss occurs, it loads the order from DB and updates cache.
func (r *Repository) GetOrderWithCache(uid string) (*model.Order, error) {
	if cached, ok := r.cache.Get(uid); ok {
		slog.Debug("cache hit", slog.String("uid", uid))
		return cached, nil
	}

	slog.Debug("cache miss", slog.String("uid", uid))
	order, err := r.getOrder(uid)
	if order != nil {
		r.cache.Set(uid, order)
		slog.Debug("cache updated after db load", slog.String("uid", uid))
	}
	return order, err
}

// saveNewOrder inserts a new order into the database.
// Associations (Delivery, Payment, Items) are also stored automatically.
func (r *Repository) saveNewOrder(tx *gorm.DB, order *model.Order) error {
	slog.Info("inserting new order", slog.String("uid", order.OrderUID))
	return tx.Create(order).Error
}

// saveExistingOrder updates an existing order and all its associations.
// FullSaveAssociations ensures Delivery, Payment, and Items are updated together.
func (r *Repository) saveExistingOrder(tx *gorm.DB, order *model.Order) error {
	slog.Info("updating existing order",
		slog.String("uid", order.OrderUID),
		slog.Int("items_count", len(order.Items)),
	)
	return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(order).Error
}

// AddOrder inserts or updates an order inside a transaction.
// After successful save, the order is refreshed in cache.
func (r *Repository) AddOrder(order *model.Order) (*model.Order, error) {
	existingOrder, err := r.getOrder(order.OrderUID)
	if err != nil {
		return nil, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if existingOrder == nil {
			return r.saveNewOrder(tx, order)
		}
		return r.saveExistingOrder(tx, order)
	})

	if err != nil {
		slog.Error("failed to save order",
			slog.String("uid", order.OrderUID),
			slog.Any("err", err),
		)
		return nil, err
	}

	// Update cache immediately with the latest order object.
	r.cache.Set(order.OrderUID, order)
	slog.Info("order saved and cached", slog.String("uid", order.OrderUID))

	return order, nil
}
