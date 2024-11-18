package database

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostgresDatabase struct {
	DatabaseConnection *gorm.DB
	Logger             *zap.SugaredLogger
	Quit               chan bool
}
