import axios from "axios";
import { Lobby, CreateLobbyData } from "@/types/lobby";

export async function fetchLobbies(): Promise<Lobby[]> {
  try {
    const response = await axios.get<Lobby[]>("/api/v1/lobbies", {
      withCredentials: true,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(
        error.response?.data?.message || "Failed to fetch lobbies",
      );
    }
    throw error;
  }
}

export async function fetchLobby(lobbyId: string): Promise<Lobby> {
  try {
    const response = await axios.get<Lobby>(`/api/v1/lobbies/${lobbyId}`, {
      withCredentials: true,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      throw new Error("Lobby not found");
    }
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || "Failed to fetch lobby");
    }
    throw error;
  }
}

export async function createLobby(data: CreateLobbyData): Promise<Lobby> {
  try {
    const response = await axios.post<Lobby>("/api/v1/lobbies", data, {
      withCredentials: true,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(
        error.response?.data?.message || "Failed to create lobby",
      );
    }
    throw error;
  }
}

export async function joinLobby(lobbyId: string): Promise<Lobby> {
  try {
    const response = await axios.post<Lobby>(
      `/api/v1/lobbies/${lobbyId}/join`,
      {},
      {
        withCredentials: true,
      },
    );
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || "Failed to join lobby");
    }
    throw error;
  }
}

export async function leaveLobby(lobbyId: string): Promise<void> {
  try {
    await axios.post(
      `/api/v1/lobbies/${lobbyId}/leave`,
      {},
      {
        withCredentials: true,
      },
    );
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || "Failed to leave lobby");
    }
    throw error;
  }
}
