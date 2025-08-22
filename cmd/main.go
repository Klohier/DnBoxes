package main

import (
	"context"
	"dango/internal/auth"
	"dango/internal/chat"

	"dango/internal/game"
	"dango/internal/session"
	"dango/internal/user"
	"dango/internal/websocket"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/redis/go-redis/v9"

	// "dango/web"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type DB struct{}

func main() {

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	dbName := os.Getenv("POSTGRES_DB")
	dbPass := os.Getenv("DATABASEPASSWORD")
	dbUser := os.Getenv("DATABASEUSER")
	dbType := os.Getenv("DATABASETYPE")
	dbHost := os.Getenv("DATABASEHOST")
	dbPort := os.Getenv("DATABASEPORT")
	port := os.Getenv("PORT")

	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", dbType, dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := pgxpool.New(context.Background(), connStr)

	slog.Info("Connecting to DB with:", "connection" , connStr)




	if err != nil {
		log.Fatal((err))
	}

	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	app := echo.New()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// SETUP for embeded react app into go binary

	// web.RegisterHandlers(app)

	// SETUP FOR Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	//SETUP CORS

	clientOrigin := os.Getenv("CLIENT_ORIGIN")
	if clientOrigin == "" {
		clientOrigin = "http://localhost:5173"
	}

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{clientOrigin, "https://localhost:4173", "https://192.168.1.42"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowCredentials: true,
	}))

	// CREATE Services and Handlers     (Must be a way to do this better?)
	userRepo := user.NewPgUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)
	loginService := auth.NewLoginService(userRepo)
	loginHandler := auth.NewLoginHandler(loginService)
	chatRepo := chat.NewPgChatRepository(db)
	chatService := chat.NewChatService(chatRepo, rdb)

	chatHandler := chat.NewChatHandler(chatService)
	gameRepo := game.NewPgGameRepository(db)
	gameService := game.NewGameService(gameRepo, rdb)
	gameHandler := game.NewGameHandler(gameService)
	sessionRepo := session.NewPgSessionRepository(db)
	sessionService := session.NewSessionService(sessionRepo)
	sessionHandler := session.NewSessionHandler(sessionService)

	handlerDeps := &websocket.HandlerDeps{
		ChatService: chatService,
		GameService: gameService,
	}

	manager := websocket.NewManager(userService, sessionService, rdb, handlerDeps)

	// Group Routes Behind a common prefix   (Possibly want to create another group that has middleware to check for authentication before accessing route?)
	api := app.Group("/api/v1")

	api.GET("/ws", manager.ServeWs)
	//Users
	api.GET("/users/:userId", userHandler.FindByID)
	api.GET("/users/me", userHandler.GetMe)
	api.POST("/users", userHandler.CreateUser)
	api.GET("/users", userHandler.GetAllUsers)

	//Auth
	api.POST("/login", loginHandler.Login)
	api.POST("/logout", loginHandler.Logout)

	//Game
	api.POST("/games", gameHandler.CreateGame)
	api.GET("/games", nil)
	api.GET("/games/:gameId/state", gameHandler.GetGameState)
	api.POST("/games/:gameId/move", gameHandler.MakeMove)
	api.POST("/games/create-bot-game", gameHandler.CreateBotGame)

	//Chat
	api.GET("/chat", chatHandler.GetAllMessageFromSession)
	api.GET("/games/:gameId/chat", chatHandler.GetAllGameMessage)
	api.POST("/games/:gameId/chat", nil)

	//Session
	api.GET("/sessions", sessionHandler.GetAllSessions)
	api.POST("/sessions", sessionHandler.CreateSession)
	api.GET("/sessions/:sessionID/chat", chatHandler.GetAllMessageFromSession)
	api.POST("/sessions/:sessionId/users/:userId", sessionHandler.AddUserToSession)
	api.DELETE("/sessions/:sessionId/users/:userId", sessionHandler.RemoveUserFromSession)

	go func() {
		log.Println("pprof server listening on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatal(err)
		}
	}()

	app.Start(fmt.Sprintf("0.0.0.0:%s", port))
	logger.Info("Server Started")
}
