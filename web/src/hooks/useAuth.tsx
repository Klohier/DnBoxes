import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
// import { useNavigate } from "react-router-dom";
import axios from "axios";
import type { User, LoginCredentials } from "../types/auth";
import { fetchUser } from "@/api/fetchUser";
// import { useNavigate } from "@tanstack/react-router";
const apiUrl =
  (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";

export function useAuth() {
  // const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Fetch current user
  const {
    data: user,
    isLoading,
    isError,
  } = useQuery<User>({
    queryKey: ["me"],
    queryFn: fetchUser,
    retry: false,
    refetchOnWindowFocus: false,
  });

  // Login mutation
  const loginMutation = useMutation<User, Error, LoginCredentials>({
    mutationFn: async (credentials) => {
      const formData = new URLSearchParams();
      formData.append("username", credentials.username);
      formData.append("password", credentials.password);

      const res = await axios.post<User>(
        `http://${apiUrl}/api/v1/login`,
        formData,
        {
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          withCredentials: true,
        }
      );
      return res.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      // void navigate("/home");
    },
  });

  // Logout mutation
  const logoutMutation = useMutation({
    mutationFn: async () => {
      await axios.post(`http://${apiUrl}/api/v1/logout`, null, {
        withCredentials: true,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      // void navigate("/");
    },
  });

  return {
    user,
    isAuthenticated: !!user,
    loading:
      isLoading ||
      loginMutation.status === "pending" ||
      logoutMutation.status === "pending",
    login: loginMutation.mutateAsync,
    logout: () => {
      logoutMutation.mutate();
    },
  };
}

export type AuthContext = ReturnType<typeof useAuth>;
