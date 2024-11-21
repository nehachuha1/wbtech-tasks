package postgres

import "gorm.io/gorm"

func MakeMigrations(conn *gorm.DB) {
	err := conn.AutoMigrate(&Order{}, &Delivery{}, &Payment{}, &Item{})
	if err != nil {
		panic("can't make migrations in postgres database")
	}
}
