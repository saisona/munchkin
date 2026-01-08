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
