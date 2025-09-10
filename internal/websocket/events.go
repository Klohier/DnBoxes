package websocket

import (
	"encoding/json"
	"log/slog"
)

// GetPlayersHandler will send a list of all currently connected players
func PlayerHandler(event Event, c *Connection) error {

	var topic string
	if err := json.Unmarshal(event.Payload, &topic); err != nil {
		slog.Error("invalid topic payload: ", "error", err)
	}

	c.manager.RLock()
	conns, ok := c.manager.subscriptions[topic]
	c.manager.RUnlock()
	if !ok {
		return nil
	}

	var players []Player
	for conn := range conns {
		players = append(players, Player{
			UserID:   conn.userID,
			Username: conn.username,
		})
	}

	c.manager.broadcast(topic, EventGetPlayers, players)

	slog.Info("Sending player list:", slog.Any("players", players))

	return nil
}

type PageEventPayload struct {
	Topic string `json:"topic"`
}

func HandlePageJoin(event Event, c *Connection) error {
	var payload PageEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	c.manager.Lock()
	if c.manager.subscriptions[payload.Topic] == nil {
		c.manager.subscriptions[payload.Topic] = make(map[*Connection]bool)
	}
	c.manager.subscriptions[payload.Topic][c] = true

	conns := c.manager.subscriptions[payload.Topic]
	usernames := make([]string, 0, len(conns))
	for conn := range conns {
		usernames = append(usernames, conn.username)
	}

	c.manager.Unlock()
	c.manager.broadcastPlayers(payload.Topic)

	slog.Info("Room members updated", "topic", payload.Topic, "members", usernames)

	return nil
}

func HandlePageLeave(event Event, c *Connection) error {
	var payload PageEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	c.manager.Lock()
	delete(c.manager.subscriptions[payload.Topic], c)
	if len(c.manager.subscriptions[payload.Topic]) == 0 {
		delete(c.manager.subscriptions, payload.Topic)
	}

	usernames := make([]string, 0, len(c.manager.subscriptions[payload.Topic]))
	for conn := range c.manager.subscriptions[payload.Topic] {
		usernames = append(usernames, conn.username)
	}
	c.manager.Unlock()

	c.manager.broadcastPlayers(payload.Topic)

	slog.Info("Room members updated after leave", "topic", payload.Topic, "members", usernames)

	return nil
}
