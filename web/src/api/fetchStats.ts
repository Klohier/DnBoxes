import axios from "axios";

export interface UserStats {
  userID: number;
  username: string;
  gamesPlayed: number;
  wins: number;
  losses: number;
  winRate: number;
  totalBoxes: number;
}

export interface LeaderboardEntry {
  rank: number;
  userID: number;
  username: string;
  value: number;
}

export interface Leaderboard {
  mostWins: LeaderboardEntry[] | null;
  mostBoxes: LeaderboardEntry[] | null;
}

export async function fetchMyStats(): Promise<UserStats> {
  const response = await axios.get<UserStats>(`/api/v1/stats/me`, {
    withCredentials: true,
  });
  return response.data;
}

export async function fetchLeaderboard(): Promise<Leaderboard> {
  const response = await axios.get<Leaderboard>(`/api/v1/stats/leaderboard`, {
    withCredentials: true,
  });
  return response.data;
}
