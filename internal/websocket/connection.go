package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	
	pingInterval = (pongWait * 9) / 10
)


type ConnectionList map[*Connection]bool


//Connection for a single websocket user
type Connection struct {
	ws     *websocket.Conn
	manager *Manager
	egress chan Event
	userID   int
	username string
	gameID *int
}

// NewConnection creates a new WebSocket connection.
func NewConnection(ws *websocket.Conn, manager *Manager, userID int, username string, gameID *int) *Connection {
	return &Connection{
		ws:     ws,
		manager: manager,
		egress:     make(chan Event, 10),
		userID: userID,
		username: username,
		gameID: gameID,
	}
}



func (c *Connection) readMessage() {
	defer func() {
		c.manager.removeConnection(c)
	}()

	c.ws.SetReadLimit(512)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := c.ws.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.ws.SetPongHandler(c.pongHandler)
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := c.ws.ReadMessage()
		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break // Break the loop to close conn & Cleanup
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
		
		}


		log.Println("got message", string(payload))
		// Route the Event
		

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handeling Message: ", err)
		}
		
	}
}



// WriteMessage sends a message to the WebSocket.
// writeMessages is a process that listens for new messages to output to the Client
func (c *Connection) writeMessage() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.removeConnection(c)
	}()

	for {
		select {
		case message, ok := <- c.egress:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				log.Println("Egress channel closed")
				if err := c.ws.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return 
			}

			

			// Write a Regular text message to the connection
			if err := c.ws.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message",)
		case <-ticker.C:
			log.Println("ping")
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return // return to break this goroutine triggeing cleanup
			}
		}

	}
}


func (c *Connection) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	log.Println("pong")
	return c.ws.SetReadDeadline(time.Now().Add(pongWait))
}