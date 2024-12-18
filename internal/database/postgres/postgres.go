package postgres

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	abstr "github.com/nehachuha1/wbtech-tasks/internal/handlers"
	pg "github.com/nehachuha1/wbtech-tasks/internal/migrations/postgres"
	dbutils "github.com/nehachuha1/wbtech-tasks/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

const (
	ErrOnCreateRow = 510 + iota
	ErrOnFindRow
)

// Управляющая структура для базы данных Postgres. Работа с БД происходит через либу gorm
type PostgresDatabase struct {
	DatabaseConnection *gorm.DB
	Logger             *zap.SugaredLogger
	Quit               chan bool
}

// Обработчик запроса на создание заказа. В нём мы декомпозируем входящий запрос на несколько сущностей
// и добавляем их в базу данных.
func (p *PostgresDatabase) CreateOrder(out chan interface{}, data []byte) {
	defer close(out)

	newOrderFromJSON := &abstr.Order{}
	queryResult := &abstr.QueryResult{
		OrderSuccess:    0,
		DeliverySuccess: 0,
		PaymentSuccess:  0,
		ItemsSuccess:    0,
		Error:           nil,
		IsSuccessQuery:  true,
	}

	err := json.Unmarshal(data, newOrderFromJSON)
	if err != nil {
		queryResult.IsSuccessQuery = false
		queryResult.Error = err
		out <- queryResult
		return
	}
	newDelivery := makeNewDelivery(newOrderFromJSON)
	newPayment := makeNewPayment(newOrderFromJSON)
	newItems, itemIDs := makeNewItems(newOrderFromJSON)
	newOrder := makeNewOrderFromJSON(newOrderFromJSON, newDelivery.DeliveryID, newPayment.PaymentID, itemIDs)

	result := p.DatabaseConnection.Table("deliveries").Create(newDelivery)
	if result.Error != nil {
		p.Logger.Warn("can't create new row in deliveries table")
		queryResult.DeliverySuccess = ErrOnCreateRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on creating new row in deliveries table")
		queryResult.IsSuccessQuery = false
	}
	result = p.DatabaseConnection.Table("payments").Create(newPayment)
	if result.Error != nil {
		p.Logger.Warn("can't create new row in payments table")
		queryResult.PaymentSuccess = ErrOnCreateRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on creating new row in payments table")
		queryResult.IsSuccessQuery = false
	}

	for _, item := range newItems {
		result = p.DatabaseConnection.Table("items").Create(item)
		if result.Error != nil {
			p.Logger.Warn(
				fmt.Sprintf("can't create new row in items table for item with ChrtId %v", item.ChrtId))
			queryResult.ItemsSuccess = ErrOnCreateRow
			queryResult.IsSuccessQuery = false
			queryResult.Error = wrapError(queryResult.Error, result.Error,
				fmt.Sprintf("can't create new row in items table for item with ChrtId %v", item.ChrtId))
		}
	}
	result = p.DatabaseConnection.Table("orders").Create(newOrder)
	if result.Error != nil {
		p.Logger.Warn("can't create new row in orders table")
		queryResult.OrderSuccess = ErrOnCreateRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on creating new row in orders table")
		queryResult.IsSuccessQuery = false
	}
	p.Logger.Info("added new order with payment, delivery and items")
	out <- queryResult
}

// Обработчик на получение заказа из БД. В нём мы "собираем" данные с сущностей постгреса в единый формат JSON,
// который представлен в описании к заданию
func (p *PostgresDatabase) GetOrder(out chan interface{}, data []byte) {
	defer close(out)

	orderFromJSON := &abstr.Order{}
	queryResult := &abstr.QueryResult{
		OrderSuccess:    0,
		DeliverySuccess: 0,
		PaymentSuccess:  0,
		ItemsSuccess:    0,
		Error:           nil,
		Data:            nil,
		IsSuccessQuery:  true,
	}

	err := json.Unmarshal(data, orderFromJSON)
	if err != nil {
		queryResult.IsSuccessQuery = false
		queryResult.Error = err
		out <- queryResult
	}
	order := &pg.Order{}
	result := p.DatabaseConnection.Table("orders").Where("order_uid = ?", orderFromJSON.OrderUid).First(
		order)
	if result.Error != nil {
		p.Logger.Warn(fmt.Sprintf("can't find order with current order_uid: %v", orderFromJSON.OrderUid))
		queryResult.OrderSuccess = ErrOnFindRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on find row in orders table")
		queryResult.IsSuccessQuery = false
		out <- queryResult
		return
	}

	delivery := &pg.Delivery{}
	result = p.DatabaseConnection.Table("deliveries").Where("delivery_id = ?",
		order.OrderDeliveryID).First(delivery)
	if result.Error != nil {
		p.Logger.Warn(fmt.Sprintf("can't find delivery with current order_delivery_id: %v",
			order.OrderDeliveryID))
		queryResult.DeliverySuccess = ErrOnFindRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on find row in deliveries table")
		queryResult.IsSuccessQuery = false
		out <- queryResult
	}
	items := make([]*pg.Item, 0)
	for _, itemID := range order.OrderItemsID {
		item := &pg.Item{}
		result = p.DatabaseConnection.Table("items").Where("chrt_id = ?", itemID).First(item)
		if result.Error != nil {
			p.Logger.Warn(fmt.Sprintf("can't find item with chrt_id: %v", itemID))
			queryResult.ItemsSuccess = ErrOnFindRow
			queryResult.Error = wrapError(queryResult.Error, result.Error,
				"failed on find row in items table")
			queryResult.IsSuccessQuery = false
		} else {
			items = append(items, item)
		}
	}
	payment := &pg.Payment{}
	result = p.DatabaseConnection.Table("payments").Where("payment_id = ?",
		order.OrderPaymentID).First(payment)
	if result.Error != nil {
		p.Logger.Warn(fmt.Sprintf("can't find item with chrt_id: %v", order.OrderPaymentID))
		queryResult.PaymentSuccess = ErrOnFindRow
		queryResult.Error = wrapError(queryResult.Error, result.Error,
			"failed on find row in payments table")
		queryResult.IsSuccessQuery = false
	}

	orderData, err := convertOrderToJSON(order, delivery, payment, items)
	if err != nil {
		p.Logger.Warn(fmt.Sprintf("marshaling to JSON order error: %v", err))
		queryResult.Error = wrapError(queryResult.Error, err, "marshaling to JSON order error")
		queryResult.Data = []byte("no data")
		queryResult.IsSuccessQuery = false
	}
	queryResult.Data = orderData
	queryResult.IsSuccessQuery = true
	out <- queryResult
}

// Метод, используемый в главной управляющей структуре DataManager для того, чтобы список всех заказов с БД.
// На выходе мы получаем слайс заказов, удовлетворяющих JSON из тех.задания, которые далее конвертируются в слайс байт
// и дальше идут на обработчик добавления данных в кэш
func (p *PostgresDatabase) GrepOrdersFromDatabase(out chan interface{}) {
	defer close(out)
	var allOrders []pg.Order

	queryResult := &abstr.QueryResult{
		Error:          nil,
		Data:           nil,
		IsSuccessQuery: false,
	}

	result := p.DatabaseConnection.Table("orders").Find(&allOrders)
	if result.Error != nil {
		p.Logger.Warn(fmt.Sprintf("failed on getting all rows in table 'orders'"))
		queryResult.Error = fmt.Errorf("failed on getting all rows in table 'orders': %v", result.Error)
	}
	fetchedOrders := make([]*abstr.Order, 0)
	for _, val := range allOrders {
		tempOrder := &abstr.Order{OrderUid: val.OrderUid}
		data, _ := json.Marshal(tempOrder)
		ch := make(chan interface{})
		go p.GetOrder(ch, data)

		orderFromDB := (<-ch).(*abstr.QueryResult)
		if orderFromDB.IsSuccessQuery && orderFromDB.Data != nil {
			tempOrder = &abstr.Order{}
			_ = json.Unmarshal(orderFromDB.Data, tempOrder)
			fetchedOrders = append(fetchedOrders, tempOrder)
		}
	}
	convertedData, err := json.Marshal(fetchedOrders)
	if err != nil {
		queryResult.IsSuccessQuery = false
		queryResult.Error = fmt.Errorf("failed on marshaling orders from DB to []byte: %v", err)
	}
	queryResult.Data = convertedData
	queryResult.IsSuccessQuery = true
	out <- queryResult
}

// Далее идут вспомогательные функции, которые используются для декомпозии и "обратной сборки"
// входящих и исходящих заказов

func makeNewDelivery(orderFromJSON *abstr.Order) *pg.Delivery {
	newDelivery := &pg.Delivery{
		DeliveryID: dbutils.GenerateNewID(),
		Name:       orderFromJSON.Delivery.Name,
		Phone:      orderFromJSON.Delivery.Phone,
		Zip:        orderFromJSON.Delivery.Zip,
		City:       orderFromJSON.Delivery.City,
		Address:    orderFromJSON.Delivery.Address,
		Region:     orderFromJSON.Delivery.Region,
		Email:      orderFromJSON.Delivery.Email,
	}
	return newDelivery
}

func makeNewPayment(orderFromJSON *abstr.Order) *pg.Payment {
	newPayment := &pg.Payment{
		PaymentID:    dbutils.GenerateNewID(),
		Transaction:  orderFromJSON.Payment.Transaction,
		RequestId:    orderFromJSON.Payment.RequestId,
		Currency:     orderFromJSON.Payment.Currency,
		Provider:     orderFromJSON.Payment.Provider,
		Amount:       orderFromJSON.Payment.Amount,
		PaymentDt:    orderFromJSON.Payment.PaymentDt,
		Bank:         orderFromJSON.Payment.Bank,
		DeliveryCost: orderFromJSON.Payment.DeliveryCost,
		GoodsTotal:   orderFromJSON.Payment.GoodsTotal,
		CustomFee:    orderFromJSON.Payment.CustomFee,
	}
	return newPayment
}

func makeNewItems(orderFromJSON *abstr.Order) ([]*pg.Item, []string) {
	newItems := make([]*pg.Item, 0)
	itemIDs := make([]string, 0)
	for _, itemFromOrder := range orderFromJSON.Items {
		newItem := &pg.Item{
			ChrtId:      itemFromOrder.ChrtId,
			TrackNumber: itemFromOrder.TrackNumber,
			Price:       itemFromOrder.Price,
			Rid:         itemFromOrder.Rid,
			Name:        itemFromOrder.Name,
			Sale:        itemFromOrder.Sale,
			Size:        itemFromOrder.Size,
			TotalPrice:  itemFromOrder.TotalPrice,
			NmId:        itemFromOrder.NmId,
			Brand:       itemFromOrder.Brand,
			Status:      itemFromOrder.Status,
		}
		newItems = append(newItems, newItem)
		itemIDs = append(itemIDs, strconv.Itoa(newItem.ChrtId))
	}
	return newItems, itemIDs
}

func makeNewOrderFromJSON(orderFromJSON *abstr.Order, deliveryID string, paymentID string, itemsID []string) *pg.Order {
	newOrder := &pg.Order{
		OrderUid:          orderFromJSON.OrderUid,
		TrackNumber:       orderFromJSON.TrackNumber,
		Entry:             orderFromJSON.Entry,
		OrderDeliveryID:   deliveryID,
		OrderPaymentID:    paymentID,
		OrderItemsID:      pq.StringArray(itemsID),
		Locale:            orderFromJSON.Locale,
		InternalSignature: orderFromJSON.InternalSignature,
		CustomerId:        orderFromJSON.CustomerId,
		DeliveryService:   orderFromJSON.DeliveryService,
		Shardkey:          orderFromJSON.Shardkey,
		SmId:              orderFromJSON.SmId,
		DateCreated:       orderFromJSON.DateCreated,
		OofShard:          orderFromJSON.OofShard,
	}
	return newOrder
}

func convertOrderToJSON(order *pg.Order, delivery *pg.Delivery, payment *pg.Payment, items []*pg.Item) ([]byte, error) {
	orderToJSON := &abstr.Order{
		OrderUid:    order.OrderUid,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,
		Delivery: struct {
			Name    string `json:"name"`
			Phone   string `json:"phone"`
			Zip     string `json:"zip"`
			City    string `json:"city"`
			Address string `json:"address"`
			Region  string `json:"region"`
			Email   string `json:"email"`
		}(struct {
			Name    string
			Phone   string
			Zip     string
			City    string
			Address string
			Region  string
			Email   string
		}{
			Name:    delivery.Name,
			Phone:   delivery.Phone,
			Zip:     delivery.Zip,
			City:    delivery.City,
			Address: delivery.Address,
			Region:  delivery.Region,
			Email:   delivery.Email,
		}),
		Payment: struct {
			Transaction  string `json:"transaction"`
			RequestId    string `json:"request_id"`
			Currency     string `json:"currency"`
			Provider     string `json:"provider"`
			Amount       int    `json:"amount"`
			PaymentDt    int    `json:"payment_dt"`
			Bank         string `json:"bank"`
			DeliveryCost int    `json:"delivery_cost"`
			GoodsTotal   int    `json:"goods_total"`
			CustomFee    int    `json:"custom_fee"`
		}(struct {
			Transaction  string
			RequestId    string
			Currency     string
			Provider     string
			Amount       int
			PaymentDt    int
			Bank         string
			DeliveryCost int
			GoodsTotal   int
			CustomFee    int
		}{
			Transaction:  payment.Transaction,
			RequestId:    payment.RequestId,
			Currency:     payment.Currency,
			Provider:     payment.Provider,
			Amount:       payment.Amount,
			PaymentDt:    payment.PaymentDt,
			Bank:         payment.Bank,
			DeliveryCost: payment.DeliveryCost,
			GoodsTotal:   payment.GoodsTotal,
			CustomFee:    payment.CustomFee,
		}),
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerId:        order.CustomerId,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmId:              order.SmId,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}

	for _, item := range items {
		itemConverted := struct {
			ChrtId      int    `json:"chrt_id"`
			TrackNumber string `json:"track_number"`
			Price       int    `json:"price"`
			Rid         string `json:"rid"`
			Name        string `json:"name"`
			Sale        int    `json:"sale"`
			Size        string `json:"size"`
			TotalPrice  int    `json:"total_price"`
			NmId        int    `json:"nm_id"`
			Brand       string `json:"brand"`
			Status      int    `json:"status"`
		}{
			ChrtId:      item.ChrtId,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmId:        item.NmId,
			Brand:       item.Brand,
			Status:      item.Status,
		}
		orderToJSON.Items = append(orderToJSON.Items, itemConverted)
	}
	data, err := json.Marshal(orderToJSON)
	if err != nil {
		return nil, fmt.Errorf("marshaling error: %v", err)
	}
	return data, nil
}

// Функция для обёртки ошибок (нужна для того, чтобы оборачивать ошибки, когда мы работаем с несколькими таблицами)
func wrapError(oldErr error, newErr error, newErrMessage string) error {
	if oldErr != nil {
		return fmt.Errorf("%s: %v | %v", newErrMessage, newErr, oldErr)
	}
	return fmt.Errorf("%v: %v", newErrMessage, newErr)
}
