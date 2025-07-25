import { User } from "@/types/auth";
import axios from "axios";
export async function fetchUser(): Promise<User | null> {
  const apiUrl =
    (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";

  try {
    const response = await axios.get<User>(`http://${apiUrl}/api/v1/users/me`, {
      withCredentials: true,
    });
    return response.data;
  } catch (err: any) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null; // not logged in
    }
    throw err; // still throw unexpected errors
  }
}
