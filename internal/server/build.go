package server

import (
	"github.com/gorilla/mux"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"github.com/nehachuha1/wbtech-tasks/internal/database"
	"github.com/nehachuha1/wbtech-tasks/internal/database/kafka/producer"
	"github.com/nehachuha1/wbtech-tasks/internal/handlers/orders"
	"go.uber.org/zap"
	"html/template"
)

func BuildNewServer(templ *template.Template, logger *zap.SugaredLogger) *mux.Router {
	kafkaConfig := config.NewKafkaConfig()
	cacheConfig := config.NewCacheConfig()
	postgresConfig := config.NewPostgresConfig()

	dataManager := database.NewDataManager(postgresConfig, cacheConfig, kafkaConfig)
	kafkaProducer := producer.NewKafkaProducer(kafkaConfig, logger)

	ordersHandler := &orders.OrderHandler{
		Templates:     templ,
		Logger:        logger,
		DataManager:   dataManager,
		KafkaProducer: kafkaProducer,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", ordersHandler.Index).Methods("GET")
	r.HandleFunc("/create", ordersHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/get", ordersHandler.GetOrder).Methods("POST")

	return r
}
