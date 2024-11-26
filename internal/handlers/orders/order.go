package orders

import (
	"encoding/json"
	"fmt"
	"github.com/nehachuha1/wbtech-tasks/internal/database"
	"github.com/nehachuha1/wbtech-tasks/internal/database/kafka/producer"
	"github.com/nehachuha1/wbtech-tasks/internal/handlers"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

type OrderHandler struct {
	Templates     *template.Template
	Logger        *zap.SugaredLogger
	DataManager   *database.DataManager
	KafkaProducer *producer.KafkaProducer
}

//func NewOrderHandler(pgCfg *config.PostgresConfig, kafkaConfig *config.KafkaConfig, cacheCfg *config.CacheConfig,
//	logger *zap.SugaredLogger) *OrderHandler {
//	return &OrderHandler{
//		Logger:        logger,
//		DataManager:   database.NewDataManager(pgCfg, cacheCfg, kafkaConfig, logger),
//		KafkaProducer: producer.NewKafkaProducer(kafkaConfig, logger),
//	}
//}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method", http.StatusBadRequest)
		return
	}
	data, err := ParseBodyToOrder(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.KafkaProducer.PushOrderToQueue(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method", http.StatusBadRequest)
		return
	}

	order := new(handlers.Order)
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, fmt.Sprintf("error by unmarshaling: %v", err), http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(order)
	if err != nil {
		http.Error(w, fmt.Sprintf("error by marshaling: %v", err), http.StatusBadRequest)
		return
	}

	out := make(chan []byte)
	go h.DataManager.RunQuery("getOrder", data, out)
	result := <-out
	w.Write(result)
}

func (h *OrderHandler) Index(w http.ResponseWriter, r *http.Request) {
	err := h.Templates.ExecuteTemplate(w, "index.html", "")
	if err != nil {
		http.Error(w, `Template error`, http.StatusInternalServerError)
		return
	}
}

func ParseBodyToOrder(r *http.Request) ([]byte, error) {
	order := new(handlers.Order)

	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		return nil, err
	}

	orderInBytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	return orderInBytes, nil
}
