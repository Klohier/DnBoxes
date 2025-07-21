import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import axios, { AxiosError } from "axios";
import type { AuthContextType, LoginCredentials } from "./types/auth";
import { useNavigate } from "react-router-dom";

type AuthProviderProps = {
  children: ReactNode;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<AxiosError | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const navigate = useNavigate();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        setLoading(true);
        await axios.get(`http://${apiUrl}/api/v1/users/me`, {
          withCredentials: true,
        });
        setIsAuthenticated(true);
      } catch (err) {
        console.log(err);
        setIsAuthenticated(false);
      } finally {
        setLoading(false);
      }
    };
    checkAuth();
  }, []);

  const login = async (credentials: LoginCredentials): Promise<void> => {
    const formData = new URLSearchParams();
    formData.append("username", credentials.username);
    formData.append("password", credentials.password);
    try {
      const response = await axios.post(
        `http://${apiUrl}/api/v1/login`,
        formData,
        {
          headers: {
            "Content-Type": "application/x-www-form-urlencoded",
          },
          withCredentials: true,
        }
      );

      if (response.status === 200) {
        setLoading(false);
        setIsAuthenticated(true);
        navigate("/home");
      }
    } catch (err) {
      const axiosError = err as AxiosError;
      const data = axiosError.response?.data as { message?: string };
      const errorMessage = data?.message || "Login failed. Please try again.";


      throw new Error(errorMessage);
      // setLoading(false);


    }
  };

  const logout = async () => {
    try {
      await axios.post(`http://${apiUrl}/api/v1/logout`, null, {
        withCredentials: true,
      });
    } catch (error) {
      console.error("Logout error:", error);
    }

    setIsAuthenticated(false);
    navigate("/");
  };

  return (
    <AuthContext.Provider
      value={{ login, logout, loading, isAuthenticated, error }}
    >
      {children}
    </AuthContext.Provider>
  );
};
