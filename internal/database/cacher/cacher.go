package cacher

import (
	"fmt"
	ch "github.com/nehachuha1/wbtech-tasks/internal/handlers"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CacheVault struct {
	data          map[string][]byte
	mu            sync.RWMutex
	ClearInterval time.Duration
	CacheLimit    int64
	Logger        *zap.SugaredLogger
	Quit          chan bool
}

func (cache *CacheVault) GetDataFromTable(key string) interface{} {
	queryResult := &ch.CacheQueryResult{
		IsSuccessQuery: false,
		Message:        "empty",
		Data:           nil,
	}

	cache.mu.RLock()
	order, isExists := cache.data[key]
	defer cache.mu.Unlock()

	if !isExists {
		queryResult.Message = fmt.Sprintf("there's no current order_id %v in database", key)
	} else {
		queryResult.Message = fmt.Sprintf("found order with order_id %v", key)
		queryResult.Data = order
		queryResult.IsSuccessQuery = true
	}
	return queryResult
}

func (cache *CacheVault) SetDataToTable(key string, value []byte) interface{} {
	queryResult := &ch.CacheQueryResult{
		Data:           nil,
		Message:        "empty",
		IsSuccessQuery: false,
	}

	cache.mu.RLock()
	if _, isExists := cache.data[key]; !isExists {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.data[key] = value
		cache.mu.Unlock()
		cache.Logger.Infow(
			fmt.Sprintf("saved order with order_id %v in cache", key),
			"source", "internal/database/cacher", "time", time.Now().String())
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("saved order with order_id %v in cache")
	} else {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.data[key] = value
		cache.mu.Unlock()

		cache.Logger.Infow(
			fmt.Sprintf("rewrite cache for order with order_id %v", key),
			"source", "internal/database/cacher", "time", time.Now().String())
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("resaved order with order_id %v in cache")
	}
	return queryResult
}
