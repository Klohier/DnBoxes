import { User } from "@/types/auth";
import axios from "axios";

export async function guestLogin(): Promise<User> {
  const response = await axios.post<User>(`/api/v1/guest`, null, {
    withCredentials: true,
  });
  return response.data;
}

export async function fetchUser(): Promise<User | null> {
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

export async function upgradeGuest(
  username: string,
  password: string,
): Promise<User> {
  const formData = new URLSearchParams();
  formData.append("username", username);
  formData.append("password", password);

  const response = await axios.post<User>(`/api/v1/users/upgrade`, formData, {
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    withCredentials: true,
  });
  return response.data;
}
