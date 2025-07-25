import axios from "axios";
import { GameStatePayload } from "@/types/websocket"; // adjust path and type as needed

export async function fetchGame(
  gameID: string
): Promise<GameStatePayload | null> {
  const apiUrl =
    (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";
  try {
    const response = await axios.get<GameStatePayload>(
      `http://${apiUrl}/api/v1/games/${gameID}/state`,
      {
        withCredentials: true,
      }
    );
    return response.data;
  } catch (error: any) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      // Game not found, return null instead of throwing
      return null;
    }
    throw error; // rethrow other errors
  }
}
