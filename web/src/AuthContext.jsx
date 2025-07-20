import { createContext, useContext, useState, useEffect } from "react";
import axios from "axios";
// import Cookies from "js-cookie";
import { useNavigate } from "react-router-dom";

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

// eslint-disable-next-line react/prop-types
export const AuthProvider = ({ children }) => {
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  // const [token, setToken] = useState(Cookies.get("DnB-Session") || null);
  const navigate = useNavigate();

  // useEffect(() => {
  //   if (token) {
  //     setIsAuthenticated(true);
  //     setLoading(false);
  //   } else {
  //     setLoading(false);
  //     setIsAuthenticated(false);
  //   }
  // }, [token]);

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

  const login = async (credentials) => {
    try {
      const response = await axios.post(
        `http://${apiUrl}/api/v1/login`,
        credentials,
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
        // setToken(Cookies.get("DnB-Session"));
        navigate("/home");
      }
    } catch (err) {
      setError(err);
      setLoading(false);
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
