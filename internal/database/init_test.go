package database

import (
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"github.com/nehachuha1/wbtech-tasks/internal/handlers"
	"github.com/nehachuha1/wbtech-tasks/pkg/log"
	"testing"
	"time"
)

func TestNewDataManager(t *testing.T) {
	pgCfg := config.NewPostgresConfig()
	cacheCfg := config.NewCacheConfig()
	logger := log.NewLogger()

	dataManager := NewDataManager(pgCfg, cacheCfg, logger)
	time.Sleep(5 * time.Second)
	dataManager.Quit <- true
}

func TestPostgresDatabase_CreateOrder(t *testing.T) {
	logger := log.NewLogger()
	pgConfig := config.NewPostgresConfig()
	cacheConfig := config.NewCacheConfig()
	dataManager := NewDataManager(pgConfig, cacheConfig, logger)

	out := make(chan interface{})
	data := []byte(`{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}
`)

	go dataManager.PostgresDB.CreateOrder(out, data)
	result := <-out
	t.Logf("Result: %#v", result)
}

func TestPostgresDatabase_GetOrder(t *testing.T) {
	logger := log.NewLogger()
	pgConfig := config.NewPostgresConfig()
	cacheConfig := config.NewCacheConfig()
	dataManager := NewDataManager(pgConfig, cacheConfig, logger)

	out := make(chan interface{})
	data := []byte(`{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "",
   "entry": "",
   "delivery": {
      "name": "",
      "phone": "",
      "zip": "",
      "city": "",
      "address": "",
      "region": "",
      "email": ""
   },
   "payment": {
      "transaction": "",
      "request_id": "",
      "currency": "",
      "provider": "",
      "amount": 0,
      "payment_dt": 0,
      "bank": "",
      "delivery_cost": 0,
      "goods_total": 0,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 0,
         "track_number": "",
         "price": 0,
         "rid": "",
         "name": "",
         "sale": 0,
         "size": "",
         "total_price": 0,
         "nm_id": 0,
         "brand": "",
         "status": 0
      }
   ],
   "locale": "",
   "internal_signature": "",
   "customer_id": "",
   "delivery_service": "",
   "shardkey": "",
   "sm_id": 0,
   "date_created": "",
   "oof_shard": ""
}
`)
	go dataManager.PostgresDB.GetOrder(out, data)
	result := (<-out).(*handlers.QueryResult)
	if result.Error != nil {
		t.Fatalf("Query error: %v", result.Error)
	}
	t.Logf("Query result bytes: %#v", result.Data)
}