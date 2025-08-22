import { createContext, useContext, ReactNode, useMemo } from "react";
import axios from "axios";
import type { LoginCredentials } from "./types/auth";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { User } from "./types/auth";
import { fetchUser } from "./api/fetchUser";
import { useRouter } from "@tanstack/react-router";
interface AuthProviderProps {
  children: ReactNode;
}

export interface AuthContextType {
  user: User | null | undefined;
  login: (credentials: LoginCredentials) => Promise<User>;
  logout: () => void;
  register: (credentials: LoginCredentials) => Promise<User>;
  loading: boolean;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const router = useRouter();

  const queryClient = useQueryClient();

  const {
    data: user,
    isLoading,
    isError,
    isFetched,
  } = useQuery<User | null>({
    queryKey: ["me"],
    queryFn: fetchUser,
    retry: false,
    refetchOnWindowFocus: false,
  });

  const loginMutation = useMutation<User, Error, LoginCredentials>({
    mutationFn: async (credentials: LoginCredentials) => {
      const formData = new URLSearchParams();
      formData.append("username", credentials.username);
      formData.append("password", credentials.password);

      const response = await axios.post<User>(`/api/v1/login`, formData, {
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        withCredentials: true,
      });
      return response.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });

  const logoutMutation = useMutation<undefined>({
    mutationFn: async () => {
      await axios.post(`/api/v1/logout`, null, {
        withCredentials: true,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      await router.navigate({ to: "/login" });
    },
  });

  const login = (credentials: LoginCredentials) =>
    loginMutation.mutateAsync(credentials);

  const logout = () => {
    logoutMutation.mutate();
  };

  const registerMutation = useMutation<User, Error, LoginCredentials>({
    mutationFn: async (credentials: LoginCredentials) => {
      const formData = new URLSearchParams();
      formData.append("username", credentials.username);
      formData.append("password", credentials.password);

      const response = await axios.post<User>(`/api/v1/users`, formData, {
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        withCredentials: true,
      });
      return response.data;
    },
    onSuccess: async (userData, credentials) => {
      try {
        console.log("Registration successful, logging in user...");

        // Automatically log in the user with the same credentials
        const loginFormData = new URLSearchParams();
        loginFormData.append("username", credentials.username);
        loginFormData.append("password", credentials.password);

        await axios.post<User>(`/api/v1/login`, loginFormData, {
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          withCredentials: true,
        });

        // Refresh user data
        await queryClient.invalidateQueries({ queryKey: ["me"] });

        console.log("Auto-login successful");

        // Optionally navigate to a specific page after registration
        // await router.navigate({ to: "/dashboard" }); // or wherever you want
      } catch (error) {
        console.error("Auto-login failed after registration:", error);

        // If auto-login fails, redirect to login page with a message
        await router.navigate({
          to: "/login",
        });
      }
    },
  });

  const register = (credentials: LoginCredentials) =>
    registerMutation.mutateAsync(credentials);

  const isAuthenticated = isFetched && !isError && !!user;
  const loading =
    (!isFetched && isLoading) ||
    loginMutation.status === "pending" ||
    logoutMutation.status === "pending" ||
    registerMutation.status === "pending";

  const authContextValue = useMemo(
    () => ({
      user,
      login,
      logout,
      register,
      loading,
      isAuthenticated,
    }),
    [user, login, logout, loading, isAuthenticated]
  );

  return (
    <AuthContext.Provider value={authContextValue}>
      {children}
    </AuthContext.Provider>
  );
};
