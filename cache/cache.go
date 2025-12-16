package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache struct {
	cache *cache.Cache
}

func New() *Cache {
	return &Cache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.cache.Set(key, value, expiration)
}

func (c *Cache) SetDefault(key string, value interface{}) {
	c.cache.Set(key, value, cache.DefaultExpiration)
}

