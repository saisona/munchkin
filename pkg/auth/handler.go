package auth

import (
	"errors"
	"net/http"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

func mapRegisterError(err error) error {
	switch {
	case errors.Is(err, ErrUsernameTaken):
		return echo.NewHTTPError(
			http.StatusConflict,
			ErrUsernameTaken,
		)
	default:
		return echo.ErrInternalServerError
	}
}

type Handler struct {
	s *Service
}

func NewAuthHandler(svc *Service) Handler {
	return Handler{s: svc}
}

// Login godoc
// @Summary Authenticate a user
// @Description Authenticate a user using username and password and return a JWT token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Authentication credentials"
// @Success 200 {object} AuthResponse "Authentication successful"
// @Failure 400 {string} string "Invalid credentials or malformed request"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	reqCtx := c.Request().Context()
	sp := trace.SpanFromContext(reqCtx)
	defer sp.End()

	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		sp.RecordError(err, trace.WithStackTrace(true))
		return echo.ErrBadRequest
	}

	token, _, err := h.s.Login(
		reqCtx,
		req.Username,
		req.Password,
	)
	if err != nil && errors.Is(err, ErrInvalidCredentials) {
		sp.RecordError(err, trace.WithStackTrace(true))
		return c.String(400, "invalid credentials")
	} else if err != nil {
		sp.RecordError(err, trace.WithStackTrace(true))
		return err
	}

	resp := AuthResponse{Token: token}
	telemetry.AuthSuccess.Inc()
	return c.JSON(200, resp)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username and password and return a JWT token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Registration credentials"
// @Success 200 {object} AuthResponse "Registration successful"
// @Failure 400 {string} string "Invalid request or user already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(c echo.Context) error {
	reqCtx := c.Request().Context()
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	token, err := h.s.Register(
		reqCtx,
		req.Username,
		req.Password,
	)
	if err != nil {
		return mapRegisterError(err)
	}

	resp := AuthResponse{Token: token}
	telemetry.AuthSuccess.Inc()
	return c.JSON(200, resp)
}

func Me(c echo.Context) error {
	return nil
}
