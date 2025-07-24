export interface GamePlayer extends Player {
  turn_order: number;
  is_spectator?: boolean;
  is_connected?: boolean;
  score: number;
}

export interface Player {
  user_id: number;
  username: string;
  avatarUrl: string;
  status: string;
}

export interface Game {
  game_id: number;
  session_id: number;
  game_name?: string;
  players: GamePlayer[];
  board_size: number;
  winner?: number;
  created_at: Date;
  current_turn?: number;
}

export interface Box {
  box_id: number;
  game_id: number;
  top_edge: boolean;
  left_edge: boolean;
  right_edge: boolean;
  bottom_edge: boolean;
  row: number;
  col: number;
  completed: boolean | null;
  completed_by: number | null;
}

export interface GameStatePayload {
  gameID?: number;
  game?: Game;
  grids?: Box[];
}

export interface WinnerPayload {
  winnerId: number;
  winnerUsername: string;
}

export interface InvitePayload {
  senderID: number;
  senderName: string;
  receiverID: number;
  receiverName: string;
  timestamp: string;
  board_size: number;
}

export interface GameStartPayload {
  gameID: string;
}

interface AcceptInvitePayload {
  playerID: number;
  senderID: number;
  board_size: number;
}

interface DeclineInvitePayload {
  inviterID: number;
}

export interface ChatMessagePayload {
  userID: number;
  username: string;
  session_id: number;
  message: string;
  timestamp: string;
}

export interface GameQuitPayload {
  gameId: number;
  playerId: number;
  session_id: number;
}

export interface GameMovePayload {
  gameId: number;
  playerId: number;
  row: number;
  col: number;
  edge: string;
}

export type Message =
  | { type: "game:move"; payload: GameMovePayload }
  | { type: "chat:new"; payload: ChatMessagePayload }
  | { type: "game:state"; payload: GameStatePayload }
  | { type: "winner_set"; payload: WinnerPayload }
  | { type: "your_turn" }
  | { type: "invalid_move" }
  | { type: "game:quit"; payload: GameQuitPayload }
  | { type: "invite:new"; payload: InvitePayload }
  | { type: "player:get"; payload?: Player[] }
  | { type: "game:new"; payload: GameStartPayload }
  | { type: "invite:accept"; payload: AcceptInvitePayload }
  | { type: "invite:decline"; payload: DeclineInvitePayload };
