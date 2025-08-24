package cache

import (
	"log/slog"
	"sync"

	model "order/internal/entity"

	lru "github.com/hashicorp/golang-lru/v2"
)

// OrderCache is a thread-safe LRU cache for storing orders
type OrderCache struct {
	mu    sync.RWMutex
	cache *lru.Cache[string, *model.Order]
}

// NewOrderCache creates a new LRU cache with the given size
func NewOrderCache(size int) (*OrderCache, error) {
	c, err := lru.New[string, *model.Order](size)
	if err != nil {
		return nil, err
	}
	return &OrderCache{cache: c}, nil
}

// Get returns the order by UID if it exists in the cache
func (oc *OrderCache) Get(uid string) (*model.Order, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	order, ok := oc.cache.Get(uid)
	if !ok {
		return nil, false
	}
	return order, true
}

// Set adds or updates an order in the cache
func (oc *OrderCache) Set(uid string, order *model.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.cache.Add(uid, order)
	slog.Debug("cache set", slog.String("uid", uid))
}

// Delete removes an order from the cache by UID
func (oc *OrderCache) Delete(uid string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.cache.Remove(uid)
	slog.Debug("cache delete", slog.String("uid", uid))
}
