// Package cache inspired from https://github.com/patrickmn/go-cache
package cache

import (
	"sync"
	"time"
)

type MemCache struct {
	items sync.Map
	ci    time.Duration
}

// NewMemCache memcache will scan all objects for every clean interval and delete expired key.
func NewMemCache(ci time.Duration) *MemCache {
	c := &MemCache{
		items: sync.Map{},
		ci:    ci,
	}

	go c.runJanitor()
	return c
}

// return true if data is fresh
func (c *MemCache) load(k string) (*Item, bool) {
	it, ok := c.get(k)
	if !ok {
		return nil, false
	}
	return it, !it.Outdated()
}

// get an item from the memcache. Returns the item or nil, and a bool indicating whether the key was found.
func (c *MemCache) get(k string) (*Item, bool) {
	tmp, ok := c.items.Load(k)
	if !ok {
		return nil, false
	}
	item := tmp.(*Item)
	if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
		return nil, false
	}
	return item, true
}

func (c *MemCache) set(k string, it *Item) {
	c.items.Store(k, it)
}

// Delete an item from the memcache. Does nothing if the key is not in the memcache.
func (c *MemCache) delete(k string) {
	c.items.Delete(k)
}

// start key scanning to delete expired keys
func (c *MemCache) runJanitor() {
	ticker := time.NewTicker(c.ci)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		}
	}
}

// DeleteExpired delete all expired items from the memcache.
func (c *MemCache) DeleteExpired() {
	c.items.Range(func(key, value interface{}) bool {
		v := value.(*Item)
		k := key.(string)
		// delete outdated for memory cache
		if v.Outdated() {
			c.items.Delete(k)
		}
		return true
	})
}
