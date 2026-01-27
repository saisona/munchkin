package auth

import (
	"strings"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(jwtKey []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// NOTE: check first for token given in query params and endpoint contains ws and looby
			isWs := strings.HasSuffix(c.Request().URL.EscapedPath(), "ws")
			if isWs {
				tokenQueryStr := c.Request().FormValue("token")

				// 3. Verify JWT
				claims, err := verifyJWT(tokenQueryStr, jwtKey)
				if err != nil {
					telemetry.AuthFailures.WithLabelValues(err.Error()).Inc()
					return echo.ErrUnauthorized
				}

				// 4. Attach identity to context
				c.Set("playerID", claims.Subject)
				return next(c)
			}

			// 1. Extract Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				telemetry.AuthFailures.WithLabelValues("missing header Authorization").Inc()
				return echo.ErrUnauthorized
			}

			// 2. Parse "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				telemetry.AuthFailures.WithLabelValues("Authorization header in bad shape").Inc()
				return echo.ErrUnauthorized
			}

			tokenStr := parts[1]

			// 3. Verify JWT
			claims, err := verifyJWT(tokenStr, jwtKey)
			if err != nil {
				telemetry.AuthFailures.WithLabelValues(err.Error()).Inc()
				return echo.ErrUnauthorized
			}

			// 4. Attach identity to context
			c.Set("playerID", claims.Subject)

			// 5. Continue request
			return next(c)
		}
	}
}
