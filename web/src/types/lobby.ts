export interface Lobby {
  lobby_id: string;
  name: string;
  host_id: number;
  player_limit: number;
  is_private: boolean;
  created_at: string;
  players?: { is_ready: boolean; userID: number }[];
}
