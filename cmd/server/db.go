package main

import (
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
)

func initDatabase(connectionString string) (*gorm.DB, error) {
	dial := postgres.New(postgres.Config{
		DSN: connectionString,
	})
	dbConn, errOpenDial := gorm.Open(dial)
	if errOpenDial != nil {
		return nil, errOpenDial
	}
	promPlug := prometheus.New(
		prometheus.Config{
			DBName:           os.Getenv("POSTGRES_DB"), // use `DBName` as metrics label
			RefreshInterval:  15,                       // Refresh metrics interval (default 15 seconds)
			StartServer:      true,                     // start http server to expose metrics
			HTTPServerPort:   8080,                     // configure http server port, default port 8080 (if you have configured multiple instances, only the first `HTTPServerPort` will be used to start server)
			MetricsCollector: []prometheus.MetricsCollector{&prometheus.Postgres{Interval: 15}},
		})

	if errPromConnect := dbConn.Use(promPlug); errPromConnect != nil {
		return nil, errPromConnect
	}
	if errOtel := dbConn.Use(otelgorm.NewPlugin(otelgorm.WithDBName("munchin"))); errOtel != nil {
		return nil, errOtel
	}

	return dbConn, nil
}
