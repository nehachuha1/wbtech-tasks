package postgres

// Интерфейс для работы с Postgres
type IPostgresDatabase interface {
	CreateOrder(out chan interface{}, data []byte)
	GetOrder(out chan interface{}, data []byte)
	GrepOrdersFromDatabase(out chan interface{})
}
