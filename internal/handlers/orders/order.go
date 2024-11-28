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

// Обработчик запроса на создание заказа. Если запрос успешно маршалится в структуру нового заказа, то
// мы пушим тело запроса в очередь топика orders и отдаём JSON о том, что запрос успешен.
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

	result := struct {
		Code    int
		Message string
	}{
		Code:    200,
		Message: "successfully created order",
	}

	data, _ = json.Marshal(result)
	w.Write(data)
}

// Обработчик запроса на получение заказа. Опять-таки пытаемся замаршалить тело запроса в структуру Order.
// Ключевым полем является для нас order_uid, по которому мы уже собираем заказ в нужный нам формат JSON и
// выводим на экран
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
	if result != nil {
		w.Write(result)
		return
	}
	message := struct {
		Code    int
		Message string
	}{
		Code:    400,
		Message: "can't find order with this order_uid",
	}
	data, _ = json.Marshal(message)
	w.Write(data)
}

// Дефолтный обработчик корневого запроса. Выводит темплейт index.html
func (h *OrderHandler) Index(w http.ResponseWriter, r *http.Request) {
	err := h.Templates.ExecuteTemplate(w, "index.html", "")
	if err != nil {
		http.Error(w, `Template error`, http.StatusInternalServerError)
		return
	}
}

// Вспомогательная функция для того, чтобы маршалить запрос в структуру Order.
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
