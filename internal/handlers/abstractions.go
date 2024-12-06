package handlers

// Структура со входящим JSON, которую мы в дальнейшем декомпозируем
type Order struct {
	OrderUid    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
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
	} `json:"payment"`
	Items []struct {
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
	} `json:"items"`
	Locale            string `json:"locale"`
	InternalSignature string `json:"internal_signature"`
	CustomerId        string `json:"customer_id"`
	DeliveryService   string `json:"delivery_service"`
	Shardkey          string `json:"shardkey"`
	SmId              int    `json:"sm_id"`
	DateCreated       string `json:"date_created"`
	OofShard          string `json:"oof_shard"`
}

// Структура, которая используется в модуле работы с Postgres.
// Ключевые поля этой структуры - IsSuccessQuery и Data.
// По булевому значению поля мы в дальнейшем проверяем, был ли успешен запрос. Если он не увенчался успехом,
// То обрабатываем ошибку из поля Error
type QueryResult struct {
	OrderSuccess    int
	DeliverySuccess int
	PaymentSuccess  int
	ItemsSuccess    int
	Error           error
	Data            []byte
	IsSuccessQuery  bool
}

// Аналогичная структура, но которая использутся в модуле работы с in memory кэшированием.
// Логика работы аналогичная
type CacheQueryResult struct {
	Data           []byte
	Message        string
	IsSuccessQuery bool
}
