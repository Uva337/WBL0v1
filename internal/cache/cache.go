package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/Uva337/WBL0v1/internal/models"
)

type Cache interface {
	Get(id string) (models.Order, bool)
	Set(id string, order models.Order)
	BulkSet(list []models.Order)
}

type GoCache struct {
	c *cache.Cache
}

func New(defaultExpiration, cleanupInterval time.Duration) *GoCache {
	return &GoCache{
		c: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (gc *GoCache) Get(id string) (models.Order, bool) {
	if v, found := gc.c.Get(id); found {
		if order, ok := v.(models.Order); ok {
			return order, true
		}
	}
	return models.Order{}, false
}

func (gc *GoCache) Set(id string, order models.Order) {
	gc.c.Set(id, order, cache.DefaultExpiration)
}

func (gc *GoCache) BulkSet(list []models.Order) {
	for _, o := range list {
		gc.Set(o.OrderUID, o)
	}
}
