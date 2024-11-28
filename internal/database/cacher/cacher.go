package cacher

import (
	"encoding/json"
	"fmt"
	ch "github.com/nehachuha1/wbtech-tasks/internal/handlers"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Управляющая структура для работы с хранилищем кэша
type CacheVault struct {
	Data          map[string][]byte
	mu            sync.RWMutex
	ClearInterval time.Duration
	CacheLimit    int64
	Logger        *zap.SugaredLogger
	Quit          chan bool
}

// Метод для получения данных по order_uid из кэша
func (cache *CacheVault) GetDataFromTable(out chan interface{}, key string) {
	defer close(out)

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

// Метод для добавления в кэш новых данных, где ключом будет order_uid
func (cache *CacheVault) SetDataToTable(out chan interface{}, data []byte) {
	defer close(out)

	queryResult := &ch.CacheQueryResult{
		Data:           nil,
		Message:        "empty",
		IsSuccessQuery: false,
	}

	order := &ch.Order{}
	err := json.Unmarshal(data, order)

	if err != nil {
		cache.Logger.Warn(fmt.Sprintf("marshaling to JSON order error: %v", err))
		queryResult.Message = fmt.Sprintf("marshaling to JSON order error: %v", err)
		out <- queryResult
	}

	cache.mu.RLock()
	if _, isExists := cache.Data[order.OrderUid]; !isExists {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.Data[order.OrderUid] = data
		cache.mu.Unlock()
		cache.Logger.Info(
			fmt.Sprintf("saved order with order_id %v in cache", order.OrderUid))
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("saved order with order_id %v in cache")
		queryResult.Data = data
		out <- queryResult

	} else {
		cache.mu.RUnlock()

		cache.mu.Lock()
		cache.Data[order.OrderUid] = data
		cache.mu.Unlock()

		cache.Logger.Info(
			fmt.Sprintf("rewrite cache for order with order_id %v", order.OrderUid))
		queryResult.IsSuccessQuery = true
		queryResult.Message = fmt.Sprintf("resaved order with order_id %v in cache")
		queryResult.Data = data
		out <- queryResult
	}
}

// Метод для подгрузки в кэш всех заказов. Используется при инициализации новой управляющей структуры для работы
// с данными
func (cache *CacheVault) LoadOrdersToCache(data []byte) {
	var allOrders []ch.Order
	err := json.Unmarshal(data, &allOrders)
	if err != nil {
		cache.Logger.Warn(fmt.Sprintf("failed on unmarshaling bytes to slice of orders: %v", err))
	}
	for _, order := range allOrders {
		out := make(chan interface{})
		marshaledOrder, _ := json.Marshal(order)
		go cache.SetDataToTable(out, marshaledOrder)
	}
	cache.Logger.Info("successfully loaded orders to cache")
}

// Метод для очистки кэша
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
			delete(cache.Data, key)
		}
	}
	queryResult := &ch.CacheQueryResult{
		Data:           nil,
		Message:        "successfully cleared cache",
		IsSuccessQuery: true,
	}
	cache.Logger.Info("cleared cache in CacheVault")
	return queryResult
}
