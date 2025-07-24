import { createContext, useContext, ReactNode } from "react";
import axios, { AxiosError } from "axios";
import type { LoginCredentials } from "./types/auth";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { User } from "./types/auth";
interface AuthProviderProps {
  children: ReactNode;
}

interface AuthContextType {
  login: (credentials: LoginCredentials) => Promise<User>;
  logout: () => void;
  loading: boolean;
  isAuthenticated: boolean;
  // error: AxiosError | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const apiUrl = (import.meta.env.VITE_API_URL as string) || "localhost:8484";
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const {
    data: user,
    isLoading,
    isError,
  } = useQuery<User>({
    queryKey: ["me", apiUrl],
    queryFn: async () => {
      const res = await axios.get<User>(`http://${apiUrl}/api/v1/users/me`, {
        withCredentials: true,
      });
      return res.data;
    },
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
      void navigate("/home");
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
      void navigate("/");
    },
  });

  const login = (credentials: LoginCredentials) =>
    loginMutation.mutateAsync(credentials);

  const logout = () => {
    logoutMutation.mutate();
  };

  const isAuthenticated = !!user && !isError;

  return (
    <AuthContext.Provider
      value={{
        login,
        logout,
        loading:
          isLoading ||
          loginMutation.status === "pending" ||
          logoutMutation.status === "pending",
        isAuthenticated,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
