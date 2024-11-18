package cacher

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CacheVault struct {
	data          map[string]interface{}
	mu            sync.RWMutex
	ClearInterval time.Duration
	CacheLimit    int64
	Logger        *zap.SugaredLogger
	Quit          chan bool
}

func (cache *CacheVault) GetData(key string) interface{} {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if val, isExists := cache.data[key]; isExists {
		return val
	}
	return nil
}

func (cache *CacheVault) SetData(key string, value interface{}) {
	cache.mu.RLock()
	if _, isExists := cache.data[key]; !isExists {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.data[key] = value
		cache.mu.Unlock()
	} else {
		cache.mu.RUnlock()
		cache.Logger.Warnw(
			fmt.Sprintf("can't save in cache value for key %v, before this key is already exists", key),
			"source", "internal/database/cacher", "time", time.Now().String())
	}
}
