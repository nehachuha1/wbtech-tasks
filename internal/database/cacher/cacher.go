package cacher

import (
	"encoding/json"
	"fmt"
	ch "github.com/nehachuha1/wbtech-tasks/internal/handlers"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CacheVault struct {
	Data          map[string][]byte
	mu            sync.RWMutex
	ClearInterval time.Duration
	CacheLimit    int64
	CurrentLength int64
	Logger        *zap.SugaredLogger
	Quit          chan bool
}

func (cache *CacheVault) GetDataFromTable(out chan interface{}, key string) {
	queryResult := &ch.CacheQueryResult{
		IsSuccessQuery: false,
		Message:        "empty",
		Data:           nil,
	}

	cache.mu.RLock()
	order, isExists := cache.Data[key]
	defer cache.mu.RUnlock()

	if !isExists {
		queryResult.Message = fmt.Sprintf("there's no current order_id %v in database", key)
	} else {
		queryResult.Message = fmt.Sprintf("found order with order_id %v", key)
		queryResult.Data = order
		queryResult.IsSuccessQuery = true
	}
	out <- queryResult
}

func (cache *CacheVault) SetDataToTable(out chan interface{}, data []byte) {
	queryResult := &ch.CacheQueryResult{
		Data:           nil,
		Message:        "empty",
		IsSuccessQuery: false,
	}

	order := &ch.Order{}
	err := json.Unmarshal(data, order)

	if err != nil {
		cache.Logger.Warnw(fmt.Sprintf("marshaling to JSON order error: %v", err),
			"source", "internal/database/cacher", "time", time.Now().String())
		queryResult.Message = fmt.Sprintf("marshaling to JSON order error: %v", err)
		out <- queryResult
	}

	cache.mu.RLock()
	if _, isExists := cache.Data[order.OrderUid]; !isExists {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.Data[order.OrderUid] = data
		cache.mu.Unlock()
		cache.Logger.Infow(
			fmt.Sprintf("saved order with order_id %v in cache", order.OrderUid),
			"source", "internal/database/cacher", "time", time.Now().String())
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("saved order with order_id %v in cache")
		queryResult.Data = data
		out <- queryResult

	} else {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.Data[order.OrderUid] = data
		cache.mu.Unlock()

		cache.Logger.Infow(
			fmt.Sprintf("rewrite cache for order with order_id %v", order.OrderUid),
			"source", "internal/database/cacher", "time", time.Now().String())
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("resaved order with order_id %v in cache")
		queryResult.Data = data
		out <- queryResult
	}
}

func (cache *CacheVault) ClearCache() interface{} {
	orderIDs := make([]string, 0)

	cache.mu.RLock()
	for key, _ := range cache.Data {
		orderIDs = append(orderIDs, key)
	}
	cache.mu.RUnlock()

	cache.mu.Lock()
	defer cache.mu.Unlock()
	for _, key := range orderIDs {
		if _, isExists := cache.Data[key]; isExists {
			cache.Logger.Infow(fmt.Sprintf("found order with order_id %v in cache", key),
				"source", "internal/database/cacher", "time", time.Now().String())
			cache.Data[key] = nil
		}
	}
	queryResult := &ch.CacheQueryResult{
		Data:           nil,
		Message:        "successfully cleared cache",
		IsSuccessQuery: true,
	}
	cache.Logger.Infow(
		"cleared cache for CacheVault",
		"source", "internal/database/cacher", "time", time.Now().String())
	return queryResult
}
