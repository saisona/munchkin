package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Healthz godoc
// @Summary Liveness probe
// @Description Returns 204 when the API process is alive.
// @Tags health
// @Success 204 "Process is alive"
// @Router /healthz [get]
func Healthz(c echo.Context) error {
	return c.NoContent(204)
}

// Startupz godoc
// @Summary Startup probe
// @Description Returns 204 when application startup has completed, otherwise 503.
// @Tags health
// @Success 204 "Application startup completed"
// @Failure 503 "Application still starting"
// @Router /startupz [get]
func Startupz(state *StartupState) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !state.IsReady() {
			return c.NoContent(503)
		}
		return c.NoContent(204)
	}
}

// Readyz godoc
// @Summary Readiness probe
// @Description Returns 204 when dependencies are ready to serve traffic, otherwise 503.
// @Tags health
// @Success 204 "Dependencies are ready"
// @Failure 503 "Dependencies are not ready"
// @Router /readyz [get]
func Readyz(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 500*time.Millisecond)
		defer cancel()

		if err := db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
			return c.NoContent(http.StatusServiceUnavailable)
		}

		return c.NoContent(http.StatusNoContent)
	}
}
