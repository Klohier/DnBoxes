package events

const (
	EventMessage       = "chat:new"
	EventSendInvite    = "invite:new"
	EventAcceptInvite  = "invite:accept"
	EventDeclineInvite = "invite:decline"
	EventQuitGame      = "game:quit"
	EventGameCreated   = "game:new"
	EventGameState     = "game:state"
	EventMakeMove      = "game:move"
	EventGetPlayers    = "player:get"
	EventNewPlayers    = "player:new"
	EventJoinPage = "page:join"
  EventLeavePage = "page:leave"
)