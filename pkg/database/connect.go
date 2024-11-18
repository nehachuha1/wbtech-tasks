package database

import (
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"github.com/nehachuha1/wbtech-tasks/internal/database/cacher"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func makeDSN(cfg *config.PostgresConfig) string {
	dsn := "postgres://" + cfg.PostgresUser + ":" + cfg.PostgresPassword + "@" +
		cfg.PostgresAddress + ":" + cfg.PostgresPort + "/" +
		cfg.PostgresDatabase
	return dsn
}

func NewPostgresDB(cfg *config.PostgresConfig, logger *zap.SugaredLogger) *PostgresDatabase {
	newDSN := makeDSN(cfg)

	dbConn, err := gorm.Open(postgres.Open(newDSN), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("can't initialize connection to postgres"))
	}

	newPostgresDatabase := &PostgresDatabase{
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

func NewCacheVault(cfg *config.CacheConfig, logger *zap.SugaredLogger) *cacher.CacheVault {
	cacheVault := &cacher.CacheVault{
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
