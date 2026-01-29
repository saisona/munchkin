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
	"syscall"
	"time"

	_ "dev.azure.com/saisona/Munchin/munchin-api/docs"
	_ "github.com/joho/godotenv/autoload"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/api"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/health"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var ErrMissingOTEL = errors.New("missing OTEL_SERVICE_NAME env")

// initTracing initializes OpenTelemetry tracing.
// If OTEL_EXPORTER_OTLP_ENDPOINT is not set, tracing is disabled gracefully.
func initTracing(ctx context.Context, svcName string) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		logger.WarnContext(ctx, "OTEL disabled: no OTEL_EXPORTER_OTLP_ENDPOINT set")
		return func(context.Context) error { return nil }, nil
	}

	if svcName == "" {
		return nil, ErrMissingOTEL
	}

	return telemetry.InitTracer(ctx, svcName, endpoint)
}

var connectionString = fmt.Sprintf(
	"host=%s user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("POSTGRES_HOST"),
	os.Getenv("POSTGRES_USER"),
	os.Getenv("POSTGRES_PASSWORD"),
	os.Getenv("POSTGRES_DB"),
)

var logger = slog.New(
	telemetry.TraceHandler{Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})},
)

// Healthz godoc
// @Summary Health probe
// @Description Liveness/readiness probe endpoint
// @Tags health
// @Success 204
// @Router /healthz [get]
func healthz(c echo.Context) error { return c.NoContent(http.StatusNoContent) }

// @title Munchin API
// @version 0.1
// @description This is the API for Munchin game backend
// @host localhost:1337
// @BasePath /
func main() {
	ctx := context.Background()

	startupState := health.NewStartupState()

	if _, exists := os.LookupEnv("JWT_SECRET"); !exists {
		logger.ErrorContext(ctx, "JWT_SECRET env variable is missing")
		os.Exit(1)
	}

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

	db, err := initDatabase(connectionString)
	if err != nil {
		panic(err)
	}

	e.Use(
		// Expose Prometheus metrics early to include all requests
		echoprometheus.NewMiddleware("munchin"),

		// Inject trace/span into request context
		otelecho.Middleware("munchin"),

		// Generate request_id header (used by logs & traces)
		middleware.RequestID(),

		// Security headers (XSS, HSTS, etc.)
		middleware.Secure(),

		// Compress responses
		middleware.Gzip(),

		// Structured request logging (OTEL-aware slog)
		middleware.RequestLogger(),

		// Cross-origin access for frontend clients
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: strings.Split(availableOrigins, ","),
		}),
	)
	e.HideBanner = true

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// adds route to serve gathered metrics
	e.GET("/metrics", echoprometheus.NewHandler())

	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	api.HandleProbeRoutes(e, db, startupState)
	api.HandleAuthRoutes(e.Group("/auth"), db)
	api.HandleLobbiesRoutes(e.Group("/lobby", auth.AuthMiddleware([]byte(jwtKey))), db)

	e.RouteNotFound("/*", func(c echo.Context) error { return c.NoContent(404) })

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
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
