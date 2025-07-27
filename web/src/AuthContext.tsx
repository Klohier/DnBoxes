import { createContext, useContext, ReactNode, useMemo } from "react";
import axios, { AxiosError } from "axios";
import type { LoginCredentials } from "./types/auth";
// import { useNavigate } from "react-router-dom";
import {
  useMutation,
  useQuery,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { User } from "./types/auth";
import { fetchUser } from "./api/fetchUser";
import { useNavigate, useRouter } from "@tanstack/react-router";
interface AuthProviderProps {
  children: ReactNode;
}

export interface AuthContextType {
  user: User | null;
  login: (credentials: LoginCredentials) => Promise<User>;
  logout: () => void;
  loading: boolean;
  isAuthenticated: boolean;
  // error: AxiosError | null;
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
  const apiUrl = (import.meta.env.VITE_API_URL as string) || "localhost:8484";
  // const router = useRouter();
  const queryClient = useQueryClient();

  const {
    data: user,
    isLoading,
    isError,
  } = useSuspenseQuery<User>({
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

      const response = await axios.post<User>(
        `http://${apiUrl}/api/v1/login`,
        formData,
        {
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          withCredentials: true,
        }
      );
      return response.data;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      // void navigate("/home");
    },
  });

  const logoutMutation = useMutation<undefined>({
    mutationFn: async () => {
      await axios.post(`http://${apiUrl}/api/v1/logout`, null, {
        withCredentials: true,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      // await router.navigate({ to: "/login" });
    },
  });

  const login = (credentials: LoginCredentials) =>
    loginMutation.mutateAsync(credentials);

  const logout = () => {
    logoutMutation.mutate();
  };

  const isAuthenticated = user && !isError;
  const loading =
    isLoading ||
    loginMutation.status === "pending" ||
    logoutMutation.status === "pending";

  const authContextValue = useMemo(
    () => ({
      user,
      login,
      logout,
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
