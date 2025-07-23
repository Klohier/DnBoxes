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
  game: Game;
  grids: Box[];
}

export interface WinnerPayload {
  winnerId: number;
  winnerUsername: string;
}

export type Message =
  | { type: "game:state"; payload: GameStatePayload }
  | { type: "winner_set"; payload: WinnerPayload }
  | { type: "your_turn" }
  | { type: "invalid_move" }
  | { type: "game:quit" };
