package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Liveness: process is alive
func Healthz(c echo.Context) error {
	return c.NoContent(204)
}

// Startup: boot completed
func Startupz(state *StartupState) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !state.IsReady() {
			return c.NoContent(503)
		}
		return c.NoContent(204)
	}
}

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
