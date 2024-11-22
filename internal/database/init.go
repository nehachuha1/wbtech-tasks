package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	cache "github.com/nehachuha1/wbtech-tasks/internal/database/cacher"
	pg "github.com/nehachuha1/wbtech-tasks/internal/database/postgres"
	"github.com/nehachuha1/wbtech-tasks/internal/handlers"
	pgmigrate "github.com/nehachuha1/wbtech-tasks/internal/migrations/postgres"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
	"time"
)

type DataManager struct {
	Logger     *zap.SugaredLogger
	postgresDB *pg.PostgresDatabase
	cacheVault *cache.CacheVault
	Commands   map[string]func(context.Context, chan interface{}, []byte)
	Quit       chan bool
	mu         sync.RWMutex
}

func NewCacheVault(cfg *config.CacheConfig, logger *zap.SugaredLogger) *cache.CacheVault {
	cacheVault := &cache.CacheVault{
		Data:          make(map[string][]byte),
		ClearInterval: cfg.ClearInterval,
		CacheLimit:    cfg.CacheLimit,
		Logger:        logger,
		Quit:          make(chan bool),
	}

	go func(cv *cache.CacheVault) {
		every := time.NewTicker(cacheVault.ClearInterval)
		select {
		case <-cacheVault.Quit:
			logger.Infow("cleared cache vault and exited",
				"source", "pkg/database/connect", "time", time.Now().String())
			return
		case <-every.C:
			logger.Infow("reloading cache from database",
				"source", "pkg/database/connect", "time", time.Now().String())
			cacheVault.CurrentLength = 0
			cacheVault.ClearCache()
		}
	}(cacheVault)

	return cacheVault
}

func makeDSN(cfg *config.PostgresConfig) string {
	dsn := "postgres://" + cfg.PostgresUser + ":" + cfg.PostgresPassword + "@" +
		cfg.PostgresAddress + ":" + cfg.PostgresPort + "/" +
		cfg.PostgresDatabase
	return dsn
}

func NewPostgresDB(cfg *config.PostgresConfig, logger *zap.SugaredLogger) *pg.PostgresDatabase {
	newDSN := makeDSN(cfg)

	dbConn, err := gorm.Open(postgres.Open(newDSN), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("can't initialize connection to postgres"))
	}

	newPostgresDatabase := &pg.PostgresDatabase{
		DatabaseConnection: dbConn,
		Logger:             logger,
		Quit:               make(chan bool),
	}

	go func() {
		<-newPostgresDatabase.Quit
		db, _ := newPostgresDatabase.DatabaseConnection.DB()
		db.Close()
		logger.Infow("connection to postgres closed", "source", "pkg/database/connect", "time", time.Now().String())
		return
	}()
	return newPostgresDatabase
}

// не используем defer, так как будем ретраить с таймаутами -> запрос может закрыть мьютекс
// для остальных
func (dm *DataManager) RunQuery(cmd string, data []byte) []byte {
	dm.mu.RLock()
	out := make(chan interface{})
	if _, isExists := dm.Commands[cmd]; isExists {
		dm.mu.RUnlock()
		dm.Logger.Infow(fmt.Sprintf("running query: %v", cmd),
			"source", "database/init/run_query", "time", time.Now().String())
		queryContext := context.Background()
		switch cmd {
		case "createOrder":
			go dm.Commands["createOrder"](queryContext, out, data)
			pgOut := (<-out).(*handlers.QueryResult)

			if pgOut.IsSuccessQuery {
				dm.Logger.Infow(fmt.Sprintf("successfully runned query: %v, trying to save result in cache",
					cmd), "source", "database/init/run_query", "time", time.Now().String())

				cacheChan := make(chan interface{})
				go dm.cacheVault.SetDataToTable(cacheChan, data)

				cacheSuccess := (<-cacheChan).(*handlers.CacheQueryResult)
				if cacheSuccess.IsSuccessQuery {
					dm.Logger.Infow("successfully saved query data in cache",
						"source", "database/init/run_query", "time", time.Now().String())
				} else {
					dm.Logger.Warnw("failed in save query data in cache",
						"source", "database/init/run_query", "time", time.Now().String())
				}
			} else {
				dm.Logger.Infow(fmt.Sprintf("failed in running query %v",
					cmd), "source", "database/init/run_query", "time", time.Now().String())
			}
			return pgOut.Data
		case "getOrder":
			currentOrder := &handlers.Order{}
			if err := json.Unmarshal(data, currentOrder); errors.Is(err, nil) {
				cacheChan := make(chan interface{})
				go dm.cacheVault.GetDataFromTable(cacheChan, currentOrder.OrderUid)
				cacheResult := (<-cacheChan).(*handlers.CacheQueryResult)
				if cacheResult.IsSuccessQuery && cacheResult.Data != nil {
					dm.Logger.Infow(fmt.Sprintf("got data from cache for order with id %v",
						currentOrder.OrderUid), "source", "database/init/run_query", "time", time.Now().String())
					return cacheResult.Data
				}
				dm.Logger.Infow(fmt.Sprintf("failed to get data from cache %v",
					currentOrder.OrderUid), "source", "database/init/run_query", "time", time.Now().String())
				go dm.Commands["getOrder"](queryContext, out, data)
				pgOut := (<-out).(*handlers.QueryResult)
				if pgOut.IsSuccessQuery && pgOut.Data != nil {
					cacheChan = make(chan interface{})
					go dm.cacheVault.SetDataToTable(cacheChan, data)
				}

				return pgOut.Data
			} else {
				dm.Logger.Infow("failed to unmarshal data from json to struct",
					"source", "database/init/run_query", "time", time.Now().String())
				return nil
			}

		default:
			return nil
		}
	} else {
		dm.Logger.Infow(fmt.Sprintf("there's no query %v in DataManager", cmd),
			"source", "database/init", "time", time.Now().String())
		dm.mu.RUnlock()
		return nil
	}
}

func (dm *DataManager) InitHandlers() {
	dm.Logger.Infow("initialized database handlers",
		"source", "pkg/database/connect", "time", time.Now().String())
	dm.Commands = map[string]func(context.Context, chan interface{}, []byte){
		"createOrder": dm.postgresDB.CreateOrder,
		"getOrder":    dm.postgresDB.GetOrder,
	}
}

func NewDataManager(pgCfg *config.PostgresConfig, cacheCfg *config.CacheConfig, logger *zap.SugaredLogger) *DataManager {
	newPostgres := NewPostgresDB(pgCfg, logger)
	newCacheVault := NewCacheVault(cacheCfg, logger)

	dataManager := &DataManager{
		Logger:     logger,
		postgresDB: newPostgres,
		cacheVault: newCacheVault,
		Quit:       make(chan bool),
	}

	dataManager.InitHandlers()
	pgmigrate.MakeMigrations(newPostgres.DatabaseConnection)

	go func() {
		select {
		case <-dataManager.Quit:
			dataManager.Logger.Infow(fmt.Sprintf("closed connection to Postgres and CacheVault"),
				"source", "database/init", "time", time.Now().String())
			dataManager.postgresDB.Quit <- true
			dataManager.cacheVault.Quit <- true
			break
		}
	}()

	return dataManager
}
