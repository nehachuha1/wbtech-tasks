package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	cache "github.com/nehachuha1/wbtech-tasks/internal/database/cacher"
	"github.com/nehachuha1/wbtech-tasks/internal/database/kafka/consumer"
	pg "github.com/nehachuha1/wbtech-tasks/internal/database/postgres"
	"github.com/nehachuha1/wbtech-tasks/internal/handlers"
	pgmigrate "github.com/nehachuha1/wbtech-tasks/internal/migrations/postgres"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
	"time"
)

// Структуру "менеджера данных". С её помощью конкурентно будем обрабатывать входящие запросы на
// сохранение/получение данных из кэша, добавление новых заказов в базу данных
type DataManager struct {
	Logger     *zap.SugaredLogger
	postgresDB *pg.PostgresDatabase
	cacheVault *cache.CacheVault
	commands   map[string]func(chan interface{}, []byte)
	Quit       chan bool
	mu         sync.RWMutex
}

// Инициализация хранилища кэша. В крутящейся горутине проверяем, не было ли сигнала на прекращение работы сервиса
func NewCacheVault(cfg *config.CacheConfig, logger *zap.SugaredLogger) *cache.CacheVault {
	cacheVault := &cache.CacheVault{
		Data:          make(map[string][]byte),
		ClearInterval: cfg.ClearInterval,
		CacheLimit:    cfg.CacheLimit,
		Logger:        logger,
		Quit:          make(chan bool),
	}

	go func() {
		for {
			select {
			case <-cacheVault.Quit:
				logger.Info("cleared cache vault and exited")
				cacheVault.ClearCache()
				return
			}
		}
	}()

	return cacheVault
}

// Функция для построения DSN (data source name) - ссылки на подключение к базе данных Postgres
func makeDSN(cfg *config.PostgresConfig) string {
	dsn := "postgres://" + cfg.PostgresUser + ":" + cfg.PostgresPassword + "@" +
		cfg.PostgresAddress + ":" + cfg.PostgresPort + "/" +
		cfg.PostgresDatabase
	return dsn
}

// Инициализация Postgres
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
		logger.Info("connection to postgres closed")
		return
	}()
	return newPostgresDatabase
}

// Функция, котороая обрабатывает все входящие запросы на данные. В её арсенале ей мапа хэндлеров кэша и постгреса.
// Через горутины она делает запросы в постгрес и кэш, логика работы следующая:
// При обработке createOrder:
// 1. Запрос в Postgres. Если запрос был успешен, то создается горутина на добавление заказа в кэш
// 2. Проверяется, успешно ли было сохранение данных в кэш
// При обработке getOrder:
// 1. Сначала идём в кэш и пытаемся получить заказ из него. При успехе возвращаем данные из кэша
// 2. Если заказ не найден в кэше, то делаем запрос в постгрес, а далее добавляем заказ в кэш
//
// Если входящая команда не может быть обработана, то мы просто возвращаем nil
func (dm *DataManager) RunQuery(cmd string, data []byte, queryOut chan []byte) {
	defer close(queryOut)
	dm.mu.RLock()
	out := make(chan interface{})
	if _, isExists := dm.commands[cmd]; isExists {
		dm.mu.RUnlock()
		dm.Logger.Info(fmt.Sprintf("running query: %v", cmd))
		switch cmd {
		case "createOrder":
			go dm.commands["createOrder"](out, data)
			pgOut := (<-out).(*handlers.QueryResult)

			if pgOut.IsSuccessQuery {
				dm.Logger.Info(fmt.Sprintf("successfully runned query: %v, trying to save result in cache", cmd))

				cacheChan := make(chan interface{})
				go dm.cacheVault.SetDataToTable(cacheChan, data)

				cacheSuccess := (<-cacheChan).(*handlers.CacheQueryResult)
				if cacheSuccess.IsSuccessQuery {
					dm.Logger.Info("successfully saved query data in cache")
				} else {
					dm.Logger.Warn("failed in save query data in cache")
				}
			} else {
				dm.Logger.Info(fmt.Sprintf("failed in running query %v", cmd))
			}
			queryOut <- pgOut.Data
			return
		case "getOrder":
			currentOrder := &handlers.Order{}
			if err := json.Unmarshal(data, currentOrder); errors.Is(err, nil) {
				cacheChan := make(chan interface{})
				go dm.cacheVault.GetDataFromTable(cacheChan, currentOrder.OrderUid)
				cacheResult := (<-cacheChan).(*handlers.CacheQueryResult)
				if cacheResult.IsSuccessQuery && cacheResult.Data != nil {
					dm.Logger.Info(fmt.Sprintf("got data from cache for order with id %v",
						currentOrder.OrderUid))
					queryOut <- cacheResult.Data
					return
				}
				dm.Logger.Info(fmt.Sprintf("failed to get data from cache for order_uid %v", currentOrder.OrderUid))
				go dm.commands["getOrder"](out, data)
				pgOut := (<-out).(*handlers.QueryResult)
				if pgOut.IsSuccessQuery && pgOut.Data != nil {
					cacheChan = make(chan interface{})
					go dm.cacheVault.SetDataToTable(cacheChan, pgOut.Data)
				}

				queryOut <- pgOut.Data
				return
			} else {
				dm.Logger.Info("failed to unmarshal data from json to struct")
				queryOut <- nil
				return
			}
		default:
			queryOut <- nil
			return
		}
	} else {
		dm.Logger.Info(fmt.Sprintf("there's no query %v in DataManager", cmd))
		dm.mu.RUnlock()
		queryOut <- nil
		return
	}
}

// Инициализация обработчик для Postgres, данная мапа используется в RunQuery
func (dm *DataManager) InitHandlers() {
	dm.Logger.Info("initialized database handlers")
	dm.commands = map[string]func(chan interface{}, []byte){
		"createOrder": dm.postgresDB.CreateOrder,
		"getOrder":    dm.postgresDB.GetOrder,
	}
}

// Инициализаия новой управляющей структуры для работы с данными. В неё грузим конфиги для Postgres, хранилища кэша
// и кафки. Внутри себя структура имеет логгер и управляющие структуры для Postgres и кэша.
// В начале инициализации мы запускам миграции. Далее грузим имеющиеся заказы в кэш.
// В горутине ловятся сообщения с кафки и в дальнешейм перенаправляются в RunQuery на создание нового заказа.
// Также в этой горутине есть тикер, который под собой имеет интервал очищения кэша. При срабатывании тикера
// делается запрос в Postgres на получение всех заказов, хранилище кэша очищается и обновляется
func NewDataManager(pgCfg *config.PostgresConfig, cacheCfg *config.CacheConfig, kafkaConfig *config.KafkaConfig,
	logger *zap.SugaredLogger) *DataManager {
	newPostgres := NewPostgresDB(pgCfg, logger)
	newCacheVault := NewCacheVault(cacheCfg, logger)

	newKafkaWorker := consumer.NewKafkaConsumer(kafkaConfig, logger)
	kafkaQuitChannel := make(chan bool)
	kafkaConsumer, _ := newKafkaWorker.InitializeConsumer(kafkaQuitChannel)

	dataManager := &DataManager{
		Logger:     logger,
		postgresDB: newPostgres,
		cacheVault: newCacheVault,
		Quit:       make(chan bool),
	}

	dataManager.InitHandlers()
	pgmigrate.MakeMigrations(newPostgres.DatabaseConnection)
	tempChannel := make(chan interface{})
	go dataManager.postgresDB.GrepOrdersFromDatabase(tempChannel)
	toLoadCacheResult := (<-tempChannel).(*handlers.QueryResult)
	if toLoadCacheResult.IsSuccessQuery && toLoadCacheResult.Data != nil {
		newCacheVault.LoadOrdersToCache(toLoadCacheResult.Data)
		logger.Info("starting pre-load orders to cache")
	}

	go func() {
		every := time.NewTicker(newCacheVault.ClearInterval)
		for {
			select {
			case inputData := <-kafkaConsumer.Messages():
				dataManager.Logger.Info(
					fmt.Sprintf("Received message in data manager from queue, starting processing"))
				returnChannel := make(chan []byte)
				go dataManager.RunQuery("createOrder", inputData.Value, returnChannel)
				result := <-returnChannel
				if result == nil {
					dataManager.Logger.Info("Successfully created new order")
				} else {
					dataManager.Logger.Info(fmt.Sprintf("failed on creating new order: %v", result))
				}
			case <-dataManager.Quit:
				dataManager.Logger.Info("closed connection to Postgres and CacheVault")
				dataManager.postgresDB.Quit <- true
				dataManager.cacheVault.Quit <- true
				kafkaQuitChannel <- true
				break
			case <-every.C:
				logger.Info("started cleaning cache")
				newCacheVault.ClearCache()
				ch := make(chan interface{})
				go dataManager.postgresDB.GrepOrdersFromDatabase(ch)
				result := (<-ch).(*handlers.QueryResult)
				if result.IsSuccessQuery && result.Data != nil {
					newCacheVault.LoadOrdersToCache(result.Data)
				}
			}
		}
	}()

	return dataManager
}
