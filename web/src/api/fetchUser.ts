import { User } from "@/types/auth";
import axios from "axios";
export async function fetchUser(): Promise<User> {
  // const apiUrl =
  //   (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";

  try {
    const response = await axios.get<User>(`/api/v1/users/me`, {
      withCredentials: true,
    });
    return response.data;
  } catch (err: unknown) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      throw new Error("User not logged in"); // not logged in
    }
    throw err; // still throw unexpected errors
  }
}
