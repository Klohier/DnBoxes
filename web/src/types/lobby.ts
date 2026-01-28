import { LobbyPlayer } from "./websocket";

export interface Lobby {
  lobby_id: string;
  name: string;
  board_size: number;
  host_id: number;
  player_limit: number;
  is_private: boolean;
  created_at: string;
  players?: LobbyPlayer[];
}

export interface CreateLobbyData {
  name: string;
  board_size: number;
  player_limit: number;
  is_private: boolean;
}
