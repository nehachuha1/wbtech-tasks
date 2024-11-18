package postgres

import (
	"gorm.io/gorm"
)

type IPostgresDatabase interface {
	CreateOrder(conn *gorm.DB, out chan interface{}, data []byte) error
	GetOrder(conn *gorm.DB, out chan interface{}, data []byte) ([]byte, error)
}
