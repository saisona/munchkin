package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/api"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	_, jwtSecretExists := os.LookupEnv("JWT_SECRET")
	if !jwtSecretExists {
		panic("JWT_SECRET env variable is missing")
	}
}

var connectionString = fmt.Sprintf(
	"host=%s user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("POSTGRES_HOST"),
	os.Getenv("POSTGRES_USER"),
	os.Getenv("POSTGRES_PASSWORD"),
	os.Getenv("POSTGRES_DB"),
)

func main() {
	telemetry.Register()

	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("munchin")) // adds middleware to gather metrics
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})))

	// Base middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(middleware.Gzip())
	e.Use(middleware.RequestLogger())

	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics

	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	db, err := initDatabase(connectionString)
	if err != nil {
		panic(err)
	}
	api.HandleAuthRoutes(e.Group("/auth"), db)
	api.HandleLobbiesRoutes(e.Group("/lobby", auth.AuthMiddleware([]byte(jwtKey))), db)

	e.RouteNotFound("/*", func(c echo.Context) error {
		return c.NoContent(404)
	})

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	go func() {
		if err := e.Start(":1337"); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	<-sigCtx.Done()
	slog.Info("Shutting down the server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Stop HTTP server (lets in-flight requests finish)
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, err.Error())
	}
}
