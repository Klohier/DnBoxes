import { AxiosError } from "axios";

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface AuthContextType {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  loading: boolean;
  isAuthenticated: boolean;
  error: AxiosError | null;
}
