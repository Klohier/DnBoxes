package events

type WebSocketSubscriber interface {
	SubscribeUser(userID int, topic string)
}
