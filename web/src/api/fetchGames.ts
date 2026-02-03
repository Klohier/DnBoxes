import axios from "axios";

export interface GamePlayer {
  user_id: number;
  username: string;
  score: number;
  turn_order: number;
}

export interface GameHistoryEntry {
  game_id: number;
  board_size: number;
  winner_id: number | null;
  created_at: string;
  ended_at: string | null;
  players: GamePlayer[];
}

export async function fetchGameHistory(): Promise<GameHistoryEntry[]> {
  const response = await axios.get<GameHistoryEntry[]>(
    `/api/v1/games/history`,
    {
      withCredentials: true,
    },
  );
  return response.data;
}
