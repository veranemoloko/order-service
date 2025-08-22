package cache

import (
	model "order/internal/entity"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// OrderCache is a thread-safe LRU cache for storing Orders
type OrderCache struct {
	mu    sync.RWMutex
	cache *lru.Cache[string, *model.Order]
}

// NewOrderCache creates a new LRU cache with a maximum of size elements
func NewOrderCache(size int) (*OrderCache, error) {
	c, err := lru.New[string, *model.Order](size)
	if err != nil {
		return nil, err
	}
	return &OrderCache{cache: c}, nil
}

// Get returns an Order by UID if it exists in the cache
func (oc *OrderCache) Get(uid string) (*model.Order, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	order, ok := oc.cache.Get(uid)
	return order, ok
}

// Set adds or updates an Order in the cache
func (oc *OrderCache) Set(uid string, order *model.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.cache.Add(uid, order)
}

// Delete removes an Order from the cache by UID
func (oc *OrderCache) Delete(uid string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.cache.Remove(uid)
}

// Keys returns a slice of all UIDs currently stored in the cache
func (oc *OrderCache) Keys() []string {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	keys := oc.cache.Keys()
	result := make([]string, len(keys))
	copy(result, keys)
	return result
}
