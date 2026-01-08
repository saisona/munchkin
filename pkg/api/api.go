package api

import (
	"os"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/lobbies"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func HandleAuthRoutes(g *echo.Group, db *gorm.DB) {
	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	dbPlayerRepo, err := auth.NewDBPlayerRepo(db)
	if err != nil {
		panic(err)
	}

	h := auth.NewAuthHandler(auth.NewService(dbPlayerRepo, auth.JwtIssuer{}, jwtKey))

	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.GET("/me", auth.Me, auth.AuthMiddleware(jwtKey))
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
	h := lobbies.NewLobbyHandler(lobbies.NewService(repo, dbPlayerRepo))
	g.POST("", h.HandleNewLobby)
	g.POST("/:id/start", h.HandleStartGame)
	g.POST("/:id/join", h.HandleJoinGame)
	g.GET("/:id/ws", h.GameWS)
}
