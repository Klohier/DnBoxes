package main

import (
	"context"
	"dango/internal/auth"
	"dango/internal/chat"
	"dango/internal/game"
	"dango/internal/infra"
	"dango/internal/lobby"
	"dango/internal/metrics"
	"dango/internal/session"
	"dango/internal/user"
	"dango/internal/websocket"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	DBName       string
	DBPass       string
	DBUser       string
	DBType       string
	DBHost       string
	DBPort       string
	Port         string
	ClientOrigin string
	TokenKey     []byte
}

type App struct {
	echo   *echo.Echo
	db     *pgxpool.Pool
	redis  *redis.Client
	logger *slog.Logger
	metrics *metrics.MetricsCollector
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration
	cfg := loadConfig()

	// Setup logger
	logger := setupLogger()
	slog.SetDefault(logger)

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		return fmt.Errorf("database setup failed: %w", err)
	}
	defer db.Close()

	// Setup Redis
	rdb := setupRedis()
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		slog.Error("Failed to flush Redis DB", "error", err)
	} else {
		slog.Info("Redis database flushed successfully")
	}

	// Create app
	app := &App{
		echo:   echo.New(),
		db:     db,
		redis:  rdb,
		logger: logger,
	}

	// Setup middleware
	app.setupMiddleware(cfg, logger)

	// Initialize services and routes
	if err := app.setupServices(cfg); err != nil {
		return fmt.Errorf("service setup failed: %w", err)
	}

	// Start pprof server
	go startPprofServer()

	// Start server
	slog.Info("Starting server", "port", cfg.Port)
	return app.echo.Start(fmt.Sprintf("0.0.0.0:%s", cfg.Port))
}

func loadConfig() *Config {
	clientOrigin := os.Getenv("CLIENT_ORIGIN")
	if clientOrigin == "" {
		clientOrigin = "http://localhost:5173"
	}

	return &Config{
		DBName:       os.Getenv("POSTGRES_DB"),
		DBPass:       os.Getenv("DATABASEPASSWORD"),
		DBUser:       os.Getenv("DATABASEUSER"),
		DBType:       os.Getenv("DATABASETYPE"),
		DBHost:       os.Getenv("DATABASEHOST"),
		DBPort:       os.Getenv("DATABASEPORT"),
		Port:         os.Getenv("PORT"),
		ClientOrigin: clientOrigin,
		TokenKey:     []byte(os.Getenv("TOKEN_KEY")),
	}
}

func setupLogger() *slog.Logger {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func setupDatabase(cfg *Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBType, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	slog.Info("Connecting to database", "host", cfg.DBHost, "database", cfg.DBName)

	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}

func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func (app *App) setupMiddleware(cfg *Config, logger *slog.Logger) {
	// Request logging
	app.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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

	// CORS
	app.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.ClientOrigin, "https://localhost:4173", "https://192.168.1.42"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowCredentials: true,
	}))
}

func (app *App) setupServices(cfg *Config) error {
	// Initialize event bus
	eventBus := infra.NewRedisEventBus(app.redis)



	app.metrics = metrics.NewMetricsCollector(context.Background(), eventBus)
	// Initialize repositories
	userRepo := user.NewPgUserRepository(app.db)
	chatRepo := chat.NewPgChatRepository(app.db)
	gameRepo := game.NewPgGameRepository(app.db)
	sessionRepo := session.NewPgSessionRepository(app.db)
	lobbyRepo := infra.NewRedisLobbyRepository(app.redis)

	// Initialize services
	userService := user.NewUserService(userRepo)
	loginService := auth.NewLoginService(userRepo)
	chatService := chat.NewChatService(chatRepo, app.redis)
	botService := game.NewBotService(eventBus)
	gameService := game.NewGameService(gameRepo, eventBus,botService)
	sessionService := session.NewSessionService(sessionRepo)
	lobbyService := lobby.NewLobbyService(lobbyRepo, eventBus)

	// Initialize handlers
	userHandler := user.NewUserHandler(userService)
	loginHandler := auth.NewLoginHandler(loginService)
	chatHandler := chat.NewChatHandler(chatService)
	sessionHandler := session.NewSessionHandler(sessionService)
	
	// Initialize WebSocket manager
	manager := websocket.NewManager(eventBus)
	gameHandler := game.NewGameHandler(gameService, manager)
	lobbyHandler := lobby.NewLobbyHandler(lobbyService, manager)
	
	go manager.Run()

	// Setup routes
	app.setupRoutes(cfg, userHandler, loginHandler, chatHandler, gameHandler, lobbyHandler, sessionHandler, manager)

	return nil
}

func (app *App) setupRoutes(
	cfg *Config,
	userHandler *user.UserHandler,
	loginHandler *auth.LoginHandler,
	chatHandler *chat.ChatHandler,
	gameHandler *game.GameHandler,
	lobbyHandler *lobby.LobbyHandler,
	sessionHandler *session.SessionHandler,
	manager *websocket.Manager,
) {
	// Public routes
	public := app.echo.Group("/api/v1")
	public.POST("/users", userHandler.CreateUser)
	public.POST("/login", loginHandler.Login)
	public.GET("/metrics", app.handleMetrics)
	
	// Protected routes
	api := app.echo.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  cfg.TokenKey,
		TokenLookup: "cookie:DnB-Session",
	}))

	// WebSocket
	api.GET("/ws", manager.ServeWs)

	// Users
	api.GET("/users/:userId", userHandler.FindByID)
	api.GET("/users/me", userHandler.GetMe)
	api.GET("/users", userHandler.GetAllUsers)

	// Lobby
	api.GET("/lobbies", lobbyHandler.GetAllLobbies)
	api.POST("/lobbies", lobbyHandler.CreateLobby)
	api.POST("/lobbies/:lobbyId/join", lobbyHandler.JoinLobby)
	api.GET("/lobbies/:lobbyId", lobbyHandler.GetLobby)
	api.POST("/lobbies/:lobbyId/ready", lobbyHandler.ToggleReady)

	// Auth
	api.POST("/logout", loginHandler.Logout)

	// Game
	api.POST("/games", gameHandler.CreateGame)
	api.GET("/games/:gameId/state", gameHandler.GetGameState)
	api.POST("/games/:gameId/move", gameHandler.MakeMove)
	api.POST("/games/create-bot-game", gameHandler.CreateBotGame)

	// Chat
	api.GET("/chat", chatHandler.GetAllMessageFromSession)
	api.GET("/games/:gameId/chat", chatHandler.GetAllGameMessage)

	// Session
	api.GET("/sessions", sessionHandler.GetAllSessions)
	api.POST("/sessions", sessionHandler.CreateSession)
	api.GET("/sessions/:sessionID/chat", chatHandler.GetAllMessageFromSession)
	api.POST("/sessions/:sessionId/users/:userId", sessionHandler.AddUserToSession)
	api.DELETE("/sessions/:sessionId/users/:userId", sessionHandler.RemoveUserFromSession)
}

func startPprofServer() {
	slog.Info("pprof server listening", "port", 6060)
	if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
		slog.Error("pprof server failed", "error", err)
	}
}

func (app *App) handleMetrics(c echo.Context) error {
	snapshot := app.metrics.GetMetrics().GetSnapshot()
	return c.JSON(http.StatusOK, snapshot)
}