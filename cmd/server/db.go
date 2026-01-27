package main

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDatabase(connectionString string) (*gorm.DB, error) {
	dial := postgres.New(postgres.Config{
		DSN: connectionString,
	})
	dbConn, errOpenDial := gorm.Open(dial)
	if errOpenDial != nil {
		return nil, errOpenDial
	}
	if errOtel := dbConn.Use(otelgorm.NewPlugin(otelgorm.WithDBName("munchin"))); errOtel != nil {
		return nil, errOtel
	}

	return dbConn, nil
}
