package database

import (
	"fmt"
	"gorm.io/gorm"
)

// Очищение базы данных от предыдущих записей (Используется строго для тестирования работы с Postgres)
func ClearDatabases(conn *gorm.DB) {
	err := conn.Exec("delete from orders where order_uid <> '';delete from items where chrt_id <> 0;" +
		"delete from payments where payment_id <> '';delete from deliveries where delivery_id <> ''")
	if err.Error != nil {
		panic(fmt.Sprintf("can't clear tables: %v", err))
	}
}
