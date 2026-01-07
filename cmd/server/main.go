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
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/api"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/log"
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
	ctx := context.Background()

	svcName := os.Getenv("OTEL_SERVICE_NAME")
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	shutdownTracer, errInitTracer := telemetry.InitTracer(ctx, svcName, otlpEndpoint)
	if errInitTracer != nil {
		panic(errInitTracer)
	}
	lpShutdown, errInitLogger := log.InitOTelLogs(ctx, svcName, otlpEndpoint)
	if errInitLogger != nil {
		panic(errInitLogger)
	}
	slog.Info("otel log pipeline initialized")

	e := echo.New()

	// Base middleware
	e.Use(
		echoprometheus.NewMiddleware("munchin"),
		otelecho.Middleware("munchin"),
		middleware.Recover(),
		middleware.RequestID(),
		middleware.Secure(),
		middleware.Gzip(),
		middleware.RequestLogger(),
	)
	e.HideBanner = true

	slog.SetDefault(slog.New(otelslog.NewHandler(svcName)))
	slog.Info(">>> slog bridge installed")

	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
	e.GET("/healthz", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

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

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	go func() {
		if err := e.Start(":1337"); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(
				sigCtx,
				"an error occured during creation of the server",
				slog.Any("error", err),
			)
		}
	}()

	<-sigCtx.Done()
	slog.Info("Shutting down the server")

	shutdownCtx, cancel := context.WithTimeout(sigCtx, 5*time.Second)
	defer cancel()

	// 1. Stop HTTP server (lets in-flight requests finish)
	if errServerShutdown := e.Shutdown(shutdownCtx); errServerShutdown != nil {
		panic(errServerShutdown)
	}
	if errLoggerShutdown := lpShutdown(ctx); errLoggerShutdown != nil {
		panic(errLoggerShutdown)
	}

	if errTracerShudown := shutdownTracer(shutdownCtx); errTracerShudown != nil {
		panic(errTracerShudown)
	}
}
