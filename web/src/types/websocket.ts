export interface GamePlayer extends Player {
  turn_order: number;

  score: number;
}

export interface Player {
  user_id: number;
  username: string;
  avatarUrl: string;
  is_anonymous: boolean;
  status: string;
}

export interface LobbyPlayer extends Player {
  is_ready: boolean;
}

export interface Game {
  game_id: number;
  // session_id: number;
  game_name?: string;
  players: GamePlayer[];
  board_size: number;
  winner_id: number | null;
  created_at: Date;
  current_turn: number;
  grid: Box[][];
}

export interface Box {
  row: number;
  col: number;
  top_edge: boolean;
  right_edge: boolean;
  bottom_edge: boolean;
  left_edge: boolean;
  owner_turn: number | null;
}

// export interface GameStatePayload {
//   // gameID?: number;
//   game: Game;
//   grids: Box[];
// }

export type GameStatePayload = Game;
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
  // session_id: number;
}

export interface GameMovePayload {
  gameId: number;
  playerId: number;
  row: number;
  col: number;
  edge: string;
}

export interface LobbyCreatedPayload {
  lobby_id: string;
  board_size: number;
  name: string;
  host_id: number;
  player_limit: number;
  is_private: boolean;
  created_at: string;
  players: LobbyPlayer[];
}

export interface LobbyUpdatedPayload {
  lobby_id: string;
  players?: LobbyPlayer[];
  status?: string;
  board_size?: number;
  host_id?: number;
  created_at?: string;
  name?: string;
  player_limit?: number;
  is_private?: boolean;
}

export interface LobbyDeletedPayload {
  lobby_id: string;
}

export type Message =
  | { type: "game:move"; payload: GameMovePayload; topic?: string }
  | { type: "chat:new"; payload: ChatMessagePayload; topic?: string }
  | { type: "game:state"; payload: GameStatePayload; topic?: string }
  | { type: "winner_set"; payload: WinnerPayload; topic?: string }
  | { type: "your_turn"; topic?: string; payload: unknown }
  | { type: "invalid_move"; topic?: string; payload: unknown }
  | { type: "game:quit"; payload: GameQuitPayload; topic?: string }
  | { type: "invite:new"; payload: InvitePayload; topic?: string }
  | { type: "player:get"; payload?: string | Player[]; topic?: string }
  | { type: "game:new"; payload: GameStartPayload; topic?: string }
  | { type: "invite:accept"; payload: AcceptInvitePayload; topic?: string }
  | { type: "invite:decline"; payload: DeclineInvitePayload; topic?: string }
  | { type: "page:join"; payload: unknown; topic?: string }
  | { type: "page:leave"; payload: unknown; topic?: string }

  // --- NEW LOBBY EVENTS ---
  | { type: "lobby_created"; payload: LobbyCreatedPayload; topic?: string }
  | { type: "lobby_updated"; payload: LobbyUpdatedPayload; topic?: string }
  | { type: "lobby_deleted"; payload: LobbyDeletedPayload; topic?: string };
