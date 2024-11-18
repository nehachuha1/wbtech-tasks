package database

import (
	"context"
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	cache "github.com/nehachuha1/wbtech-tasks/internal/database/cacher"
	db "github.com/nehachuha1/wbtech-tasks/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
	"time"
)

type DataManager struct {
	Logger     *zap.SugaredLogger
	PostgresDB *db.PostgresDatabase
	CacheVault *cache.CacheVault
	Commands   map[string]func(*gorm.DB, context.Context, chan interface{}, []byte)
	Quit       chan bool
	mu         sync.RWMutex
}

// не используем defer, так как будем ретраить с таймаутами -> запрос может закрыть мьютекс
// для остальных
func (dm *DataManager) RunQuery(cmd string, out chan interface{}, data []byte) {
	dm.mu.RLock()
	if _, isExists := dm.Commands[cmd]; isExists {
		dm.mu.RUnlock()
		dm.Logger.Infow(fmt.Sprintf("running query: %v", cmd),
			"source", "database/init", "time", time.Now().String())
		ctx := context.Background()
		switch cmd {
		case "createOrder":
			dm.Commands[cmd](dm.PostgresDB.DatabaseConnection, ctx, out, data)
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
	//dm.Commands = map[string]func(*gorm.DB, context.Context, chan interface{}, []byte){
	//	"",
	//}
}

func NewDataManager(pgCfg *config.PostgresConfig, cacheCfg *config.CacheConfig, logger *zap.SugaredLogger) *DataManager {
	newPostgres := db.NewPostgresDB(pgCfg, logger)
	newCacheVault := db.NewCacheVault(cacheCfg, logger)

	dataManager := &DataManager{
		Logger:     logger,
		PostgresDB: newPostgres,
		CacheVault: newCacheVault,
		Quit:       make(chan bool),
	}

	dataManager.InitHandlers()

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
