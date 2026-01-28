package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "dev.azure.com/saisona/Munchin/munchin-api/docs"
	_ "github.com/joho/godotenv/autoload"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

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

func initTracing(ctx context.Context, svcName string) (func(context.Context) error, error) {
	otlpEndpoint, tracingOTELEndpoint := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !tracingOTELEndpoint {
		logger.Warn("OTEL cannot be set as no OTEL_EXPORTER_OTLP_ENDPOINT env found")
	} else if tracingOTELEndpoint && svcName == "" {
		logger.Warn("OTEL is set but OTEL_SERVICE_NAME env is missing")
		return nil, errors.New("missing OTEL_SERVICE_NAME env")
	}

	shutdownTracer, errInitTracer := telemetry.InitTracer(ctx, svcName, otlpEndpoint)
	if errInitTracer != nil {
		return nil, errInitTracer
	}
	return shutdownTracer, nil
}

var connectionString = fmt.Sprintf(
	"host=%s user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("POSTGRES_HOST"),
	os.Getenv("POSTGRES_USER"),
	os.Getenv("POSTGRES_PASSWORD"),
	os.Getenv("POSTGRES_DB"),
)

var (
	_jsonLogger = slog.NewJSONHandler(os.Stdout, nil)
	logger      = slog.New(telemetry.TraceHandler{Handler: _jsonLogger})
)

// GetLobbies godoc
// @Summary Check health of the service
// @Description Works as a probe like healhtz check
// @Tags health
// @Success 204
// @Router /healthz [get]
func healhtz(c echo.Context) error { return c.NoContent(http.StatusNoContent) }

// @title Munchin API
// @version 0.1
// @description This is the API for Munchin game backend
// @host localhost:1337
// @BasePath /
func main() {
	ctx := context.Background()

	shutdownTracer, errInitTracing := initTracing(ctx, os.Getenv("OTEL_SERVICE_NAME"))
	if errInitTracing != nil {
		panic(errInitTracing)
	}

	// 3. NOW tracing (traces don't affect slog)
	telemetry.Register()

	// 4. Everything else AFTER
	e := echo.New()
	availableOrigins := os.Getenv("MUNCHIN_ALLOWED_ORIGINS")
	slog.SetDefault(logger)

	e.Use(
		echoprometheus.NewMiddleware("munchin"),
		otelecho.Middleware("munchin"),
		middleware.RequestID(),
		middleware.Secure(),
		middleware.Gzip(),
		middleware.RequestLogger(), // ← NOW it uses OTEL slog
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: strings.Split(availableOrigins, ","),
		}),
	)
	e.HideBanner = true

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// adds route to serve gathered metrics
	e.GET("/metrics", echoprometheus.NewHandler())

	e.GET("/healthz", healhtz)

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
			slog.With(slog.Any("error", err)).
				ErrorContext(sigCtx, "an error occured during creation of the server")
		}
	}()

	<-sigCtx.Done()
	logger.InfoContext(sigCtx, "Shutting down the server")

	shutdownCtx, cancel := context.WithTimeout(sigCtx, 5*time.Second)
	defer cancel()

	// 1. Stop HTTP server (lets in-flight requests finish)
	if errServerShutdown := e.Shutdown(shutdownCtx); errServerShutdown != nil {
		logger.With(slog.Any("error", errServerShutdown)).
			ErrorContext(shutdownCtx, "Shutting down the server failed")
	}
	if errTracerShudown := shutdownTracer(shutdownCtx); errTracerShudown != nil {
		logger.With(slog.Any("error", errTracerShudown)).
			ErrorContext(shutdownCtx, "Shutting down the server failed")
	}
}
