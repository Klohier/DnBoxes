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

// Domain event types matching backend event sourcing

export interface DomainEvent {
  type: string;
  occurred_at: string;
  aggregate_id: string;
  version: number;
  payload: unknown;
}

export interface GameCreatedPayload {
  game_id: number;
  board_size: number;
  players: GamePlayer[];
}

export interface MoveAppliedPayload {
  turn_order: number;
  row: number;
  col: number;
  edge: string;
}

export interface BoxCompletedPayload {
  row: number;
  col: number;
  owner_turn: number;
}

export interface TurnPassedPayload {
  next_turn: number;
}

export interface GameEndedPayload {
  winner_id: number | null;
}

export interface GameForfeitedPayload {
  forfeited_by: number;
  winner_id: number | null;
}

export async function fetchGameEvents(
  gameID: number,
): Promise<DomainEvent[]> {
  const response = await axios.get<DomainEvent[]>(
    `/api/v1/games/${gameID}/events`,
    {
      withCredentials: true,
    },
  );
  return response.data;
}
