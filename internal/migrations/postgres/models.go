package postgres

import (
	"github.com/lib/pq"
)

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
