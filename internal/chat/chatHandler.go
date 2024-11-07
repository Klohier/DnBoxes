package chat

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)


type ChatHandler struct {
    chatService *ChatService
	logger *slog.Logger
}
//  
func NewChatHandler(chatService *ChatService) *ChatHandler{
	return &ChatHandler{
	chatService: chatService,	
	logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),}
}



   var clients = make(map[*websocket.Conn]struct{})



// Upgrades Handle into a WebSocket Handle and Starts goroutine for handling clients
func (h *ChatHandler) ServeWs(c echo.Context) error {

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
   }

	ctx := context.Background()
    conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        fmt.Println("Upgrade error:", err)
        return err
    }

	go h.handleClient(ctx, conn)
    // defer conn.Close()

	return err

}


func (h *ChatHandler) handleClient(ctx context.Context,c *websocket.Conn){
	defer func() {
		delete(clients, c)
		log.Println("Closing websocket")
		c.Close()
	}()
	clients[c] = struct{}{}

	// Infinite loop that sends message to all clients
	//Binds incoming Message into Message Struct and sets timestamp
	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading Message: %v", err)
			return
		}
		msg.TimeStamp = time.Now().UTC()
		err = h.chatService.SendMessage(ctx, msg)
		if err != nil {
            log.Printf("Error sending message: %v", err)
            return
        }
		broadcast(msg)
	}

}

// Sends Message to Clients
func broadcast(msg Message) {
    for conn := range clients {
        conn.WriteJSON(msg)
    }
}