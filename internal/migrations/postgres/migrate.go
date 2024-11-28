package postgres

import "gorm.io/gorm"

// Библиотека gorm предоставляет возможность делать автоматические миграции нужных нам сущностей
// Если сущность не была создана, то gorm её автоматически создаст
func MakeMigrations(conn *gorm.DB) {
	err := conn.AutoMigrate(&Order{}, &Delivery{}, &Payment{}, &Item{})
	if err != nil {
		panic("can't make migrations in postgres database")
	}
}
