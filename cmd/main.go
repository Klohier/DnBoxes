package main

import (
	"context"
	"dango/internal/auth"
	"dango/internal/chat"
	"dango/internal/websocket"
	"fmt"

	"dango/internal/game"
	"dango/internal/user"

	"dango/web"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)



type DB struct {}


func main() {
	
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbName := os.Getenv("DATABASENAME")
	dbPass := os.Getenv("DATABASEPASSWORD")
	dbUser := os.Getenv("DATABASEUSER")
	dbType := os.Getenv("DATABASETYPE")
	dbHost := os.Getenv("DATABASEHOST")
	dbPort := os.Getenv("DATABASEPORT")
	port := os.Getenv("PORT")


	// Should Move into a .env file soon!!!!!!!

	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", dbName,dbUser,dbPass, dbHost, dbPort , dbType)
	
	db, err := pgxpool.Connect(context.Background(), connStr)
	
	
	if err != nil {
		log.Fatal((err))
	}

	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}




	app := echo.New()

	// SETUP for embeded react app into go binary

	web.RegisterHandlers(app)

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

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"http://localhost:5173"},
        AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowCredentials: true,
    }))

	// CREATE Services and Handlers     (Must be a way to do this better?)
	userRepo := user.NewPgUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)
	loginService := auth.NewLoginService(userRepo)
	loginHandler := auth.NewLoginHandler(loginService)
	chatRepo := chat.NewPgChatRepository(db)
	chatService := chat.NewChatService(chatRepo)
	chatHandler := chat.NewChatHandler(chatService)
	gameRepo := game.NewPgGameRepository(db)
	gameService := game.NewGameService(gameRepo)
	gameHandler := game.NewGameHandler(gameService)

	manager := websocket.NewManager(gameService, userService, chatService)

	// messageRouter := websocket.NewMessageRouter()

	// Register handlers for different message types.
	// messageRouter.RegisterHandler("chat", chatService.HandleChatMessage)
	// messageRouter.RegisterHandler("invite", gameService.HandleGameInvite)
	// messageRouter.RegisterHandler("move", gameService.HandleGameMove)
	
	// Create the WebSocketHandler with the router.
	// websocketHandler := websocket.NewWebSocketHandler(messageRouter)
	
	// Register WebSocket endpoint.
	

	


	// Group Routes Behind a common prefix   (Possibly want to create another group that has middleware to check for authentication before accessing route?)
	api := app.Group("/api/v1")




api.GET("/ws", manager.ServeWs)
	//Users
	api.GET("/users/:userId" , userHandler.FindByID)
	api.POST("/users", userHandler.CreateUser)
	api.GET("/users", userHandler.GetAllUsers)

	api.POST("/login", loginHandler.Login)

	//Game
	api.POST("/games", gameHandler.CreateGame)
	api.GET("/games", nil)
	api.GET("/games/:gameId/grid", gameHandler.GetGrids)
	api.POST("/games/:gameId/move", gameHandler.MakeMove)

	//Chat
	api.GET("/chat", chatHandler.GetAllMessage)
	// api.GET("/chathistory", chatHandler.GetChatHistory)
	api.GET("/games/:gameId/chat", chatHandler.GetAllGameMessage)
	api.POST("/games/:gameId/chat", nil)

	
	
	app.Start(fmt.Sprintf("0.0.0.0:%s", port))
	logger.Info("Server Started")
}