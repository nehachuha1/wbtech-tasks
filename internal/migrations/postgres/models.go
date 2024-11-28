package postgres

import (
	"github.com/lib/pq"
)

// Файл с внутренними сущностями сервиса

// Структура сущности заказа. В ней переопределены поля с доставкой, списком вещей и платежом.
// В дальнейшем при поступлении нового JSON в переопределенные поля присваиваются новые ID,
// по ним сервис ищет в нескольких таблицах информацию и собирает в обратный JSON
type Order struct {
	OrderUid          string
	TrackNumber       string
	Entry             string
	OrderDeliveryID   string
	OrderPaymentID    string
	OrderItemsID      pq.StringArray `gorm:"type:text[]"`
	Locale            string
	InternalSignature string
	CustomerId        string
	DeliveryService   string
	Shardkey          string
	SmId              int
	DateCreated       string
	OofShard          string
}

// Сущность для доставки
type Delivery struct {
	DeliveryID string
	Name       string
	Phone      string
	Zip        string
	City       string
	Address    string
	Region     string
	Email      string
}

// Сущность для платежа
type Payment struct {
	PaymentID    string
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
}

// Сущность для вещи из заказа. В таблице orders хранится массив из chrt_id
type Item struct {
	ChrtId      int
	TrackNumber string
	Price       int
	Rid         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmId        int
	Brand       string
	Status      int
}
