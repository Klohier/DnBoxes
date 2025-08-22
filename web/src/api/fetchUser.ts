import { User } from "@/types/auth";
import axios from "axios";
export async function fetchUser(): Promise<User | null> {
  // const apiUrl =
  //   (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";

  try {
    const response = await axios.get<User>(`/api/v1/users/me`, {
      withCredentials: true,
    });
    return response.data;
  } catch (err: unknown) {
    if (axios.isAxiosError(err)) {
      const status = err.response?.status;
      if (status === 401 || status === 403 || status === 404) {
        return null;
      }
    }
    throw err;
  }
}
