package api

import (
	"os"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/game"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/health"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/lobbies"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func HandleProbeRoutes(e *echo.Echo, db *gorm.DB, startupState *health.StartupState) {
	e.GET("/healthz", health.Healthz)
	e.GET("/startupz", health.Startupz(startupState))
	e.GET("/readyz", health.Readyz(db))
}

func HandleAuthRoutes(g *echo.Group, db *gorm.DB) {
	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	dbPlayerRepo, err := auth.NewDBPlayerRepo(db)
	if err != nil {
		panic(err)
	}

	h := auth.NewAuthHandler(auth.NewService(dbPlayerRepo, auth.JwtIssuer{}, jwtKey))

	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
}

func HandleLobbiesRoutes(g *echo.Group, db *gorm.DB) {
	repo, err := lobbies.NewDBLobbyRepo(db)
	if err != nil {
		panic(err)
	}

	dbPlayerRepo, err := auth.NewDBPlayerRepo(db)
	if err != nil {
		panic(err)
	}
	h := lobbies.NewLobbyHandler(lobbies.NewService(repo, dbPlayerRepo), game.NewGameHub())

	g.POST("", h.HandleNewLobby)
	g.GET("", h.ListLobbies)
	g.GET("model", h.GetAllLobbies)

	handleLobbyGroup(g.Group("/:id"), h)
}

func handleLobbyGroup(g *echo.Group, handler lobbies.Handler) {
	g.POST("/start", handler.HandleStartGame)
	g.POST("/join", handler.HandleJoinGame)
	g.GET("/ws", handler.GameWS)
}
