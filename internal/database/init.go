package database

import (
	"context"
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	cache "github.com/nehachuha1/wbtech-tasks/internal/database/cacher"
	pg "github.com/nehachuha1/wbtech-tasks/internal/database/postgres"
	pgmigrate "github.com/nehachuha1/wbtech-tasks/internal/migrations/postgres"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
	"time"
)

type DataManager struct {
	Logger     *zap.SugaredLogger
	PostgresDB *pg.PostgresDatabase
	CacheVault *cache.CacheVault
	Commands   map[string]func(context.Context, chan interface{}, []byte)
	Quit       chan bool
	mu         sync.RWMutex
}

func NewCacheVault(cfg *config.CacheConfig, logger *zap.SugaredLogger) *cache.CacheVault {
	cacheVault := &cache.CacheVault{
		ClearInterval: cfg.ClearInterval,
		CacheLimit:    cfg.CacheLimit,
		Logger:        logger,
		Quit:          make(chan bool),
	}

	// TODO: Move reload cache to database/init.go DatabaseManager
	go func() {
		//cleanTimer := time.NewTicker(cacheVault.ClearInterval)
		select {
		case <-cacheVault.Quit:
			logger.Infow("cleared cache vault",
				"source", "pkg/database/connect", "time", time.Now().String())
			return
			//case <-cleanTimer.C:
			//	//go cacheVault.ReloadCache()
			//	logger.Infow("started reloading cache vault after 1 hour...",
			//		"source", "pkg/database/connect", "time", time.Now().String())
		}
	}()

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
	}()
	return newPostgresDatabase
}

// не используем defer, так как будем ретраить с таймаутами -> запрос может закрыть мьютекс
// для остальных
func (dm *DataManager) RunQuery(cmd string, out chan interface{}, data []byte) {
	dm.mu.RLock()
	if _, isExists := dm.Commands[cmd]; isExists {
		dm.mu.RUnlock()
		dm.Logger.Infow(fmt.Sprintf("running query: %v", cmd),
			"source", "database/init", "time", time.Now().String())
		queryContext := context.Background()
		switch cmd {
		case "createOrder":
			dm.Commands["createOrder"](queryContext, out, data)
		case "getOrder":
			dm.Commands["getOrder"](queryContext, out, data)
		}
	} else {
		dm.Logger.Infow(fmt.Sprintf("there's no query %v in DataManager", cmd),
			"source", "database/init", "time", time.Now().String())
		dm.mu.RUnlock()
	}
}

func (dm *DataManager) InitHandlers() {
	dm.Logger.Infow("initialized database handlers",
		"source", "pkg/database/connect", "time", time.Now().String())
	dm.Commands = map[string]func(context.Context, chan interface{}, []byte){
		"createOrder": dm.PostgresDB.CreateOrder,
		"getOrder":    dm.PostgresDB.GetOrder,
	}
}

func NewDataManager(pgCfg *config.PostgresConfig, cacheCfg *config.CacheConfig, logger *zap.SugaredLogger) *DataManager {
	newPostgres := NewPostgresDB(pgCfg, logger)
	newCacheVault := NewCacheVault(cacheCfg, logger)

	dataManager := &DataManager{
		Logger:     logger,
		PostgresDB: newPostgres,
		CacheVault: newCacheVault,
		Quit:       make(chan bool),
	}

	dataManager.InitHandlers()
	pgmigrate.MakeMigrations(newPostgres.DatabaseConnection)

	go func() {
		select {
		case <-dataManager.Quit:
			dataManager.Logger.Infow(fmt.Sprintf("closed connection to Postgres and CacheVault"),
				"source", "database/init", "time", time.Now().String())
			dataManager.PostgresDB.Quit <- true
			dataManager.CacheVault.Quit <- true
			break
		}
	}()

	return dataManager
}
