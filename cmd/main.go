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
	"dango/internal/stats"
	"dango/internal/user"
	"dango/internal/websocket"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type Config struct {
	DBName       string
	DBPass       string
	DBUser       string
	DBType       string
	DBHost       string
	DBPort       string
	Port         string
	RedisPassword string
	RedisAddr    string
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
	rdb := setupRedis(cfg)
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
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
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

func setupRedis(cfg *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
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
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderXCSRFToken},
		AllowCredentials: true,
	}))

	// CSRF protection
	app.echo.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:" + echo.HeaderXCSRFToken + ",form:_csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieSameSite: http.SameSiteNoneMode,
		CookieHTTPOnly: false,
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
	timerService := game.NewGameTimerService(eventBus)
	gameService := game.NewGameService(gameRepo, eventBus, botService, timerService)
	sessionService := session.NewSessionService(sessionRepo)
	lobbyService := lobby.NewLobbyService(lobbyRepo, eventBus)

	// Wire timer forfeit callback (avoids circular dependency)
	timerService.SetForfeitFunc(func(ctx context.Context, gameID, playerID int) error {
		_, err := gameService.ForfeitGame(ctx, gameID, playerID)
		return err
	})

	// Start listening for connect/disconnect events for timer
	go timerService.ListenForConnectionEvents()

	// Initialize stats
	statsRepo := stats.NewPgStatsRepository(app.db)
	statsService := stats.NewStatsService(statsRepo)
	statsHandler := stats.NewStatsHandler(statsService)

	// Initialize handlers
	userHandler := user.NewUserHandler(userService)
	loginHandler := auth.NewLoginHandler(loginService, userService)
	chatHandler := chat.NewChatHandler(chatService)
	sessionHandler := session.NewSessionHandler(sessionService)

	// Initialize WebSocket manager
	manager := websocket.NewManager(eventBus)
	gameHandler := game.NewGameHandler(gameService, timerService, manager)
	lobbyHandler := lobby.NewLobbyHandler(lobbyService, manager)

	// Start chat persistence worker (subscribes to EventBus independently)
	chatService.StartPersistenceWorker(eventBus)

	go manager.Run()

	// Setup routes
	app.setupRoutes(cfg, userHandler, loginHandler, chatHandler, gameHandler, lobbyHandler, sessionHandler, statsHandler, manager)

	return nil
}

// newAuthRateLimiter creates rate limiting middleware for auth endpoints.
// rateInterval is the time between allowed requests; burst is the max burst size.
func newAuthRateLimiter(rateInterval time.Duration, burst int) echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Every(rateInterval),
			Burst:     burst,
			ExpiresIn: 3 * time.Minute,
		},
	)
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]string{
				"message": "too many requests, please try again later",
			})
		},
	})
}

func (app *App) setupRoutes(
	cfg *Config,
	userHandler *user.UserHandler,
	loginHandler *auth.LoginHandler,
	chatHandler *chat.ChatHandler,
	gameHandler *game.GameHandler,
	lobbyHandler *lobby.LobbyHandler,
	sessionHandler *session.SessionHandler,
	statsHandler *stats.StatsHandler,
	manager *websocket.Manager,
) {
	// Public routes with rate limiting
	public := app.echo.Group("/api/v1")
	public.POST("/login", loginHandler.Login, newAuthRateLimiter(12*time.Second, 5))        // ~5 per minute
	public.POST("/users", userHandler.CreateUser, newAuthRateLimiter(20*time.Second, 3))     // ~3 per minute
	public.POST("/guest", loginHandler.GuestLogin, newAuthRateLimiter(6*time.Second, 5))     // ~10 per minute
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
	api.POST("/lobbies/:lobbyId/leave", lobbyHandler.LeaveLobby)
	api.DELETE("/lobbies/:lobbyId", lobbyHandler.DeleteLobby)

	// Auth
	api.POST("/logout", loginHandler.Logout)

	// Users (upgrade)
	api.POST("/users/upgrade", userHandler.UpgradeGuest)

	// Game
	api.POST("/games", gameHandler.CreateGame)
	api.GET("/games/history", gameHandler.GetGameHistory)
	api.GET("/games/:gameId/state", gameHandler.GetGameState)
	api.POST("/games/:gameId/move", gameHandler.MakeMove)
	api.POST("/games/:gameId/forfeit", gameHandler.ForfeitGame)
	api.POST("/games/create-bot-game", gameHandler.CreateBotGame)
	api.GET("/games/:gameId/timer", gameHandler.GetTimerState)
	api.GET("/games/:gameId/events", gameHandler.GetGameEvents)

	// Chat (rate limited to prevent excessive polling)
	chatRateLimiter := newAuthRateLimiter(2*time.Second, 10) // ~30 per minute, burst 10
	api.GET("/chat", chatHandler.GetGlobalMessages, chatRateLimiter)
	api.GET("/games/:gameId/chat", chatHandler.GetAllGameMessage, chatRateLimiter)

	// Stats
	api.GET("/stats/me", statsHandler.GetMyStats)
	api.GET("/stats/users/:userId", statsHandler.GetUserStats)
	api.GET("/stats/leaderboard", statsHandler.GetLeaderboard)

	// Session
	api.GET("/sessions", sessionHandler.GetAllSessions)
	api.POST("/sessions", sessionHandler.CreateSession)
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