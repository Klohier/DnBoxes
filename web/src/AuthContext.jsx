import { createContext, useContext, useState, useEffect } from "react";
import axios from "axios";
import Cookies from "js-cookie";
import { useNavigate } from "react-router-dom";

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

// eslint-disable-next-line react/prop-types
export const AuthProvider = ({ children }) => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [token, setToken] = useState(Cookies.get("DnB-Session") || null);
  const navigate = useNavigate();

  useEffect(() => {
    if (token) {
      setIsAuthenticated(true);
      setLoading(false);
    } else {
      setLoading(false);
      setIsAuthenticated(false);
    }
  }, [token]);

  const login = async (credentials) => {
    try {
      const response = await axios.post(
        "http://localhost:8484/api/v1/login",
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
        setToken(Cookies.get("DnB-Session"));
        navigate("/home");
      }
    } catch (err) {
      setError(err);
      setLoading(false);
    }
  };

  const logout = () => {
    Cookies.remove("DnB-Session");
    setIsAuthenticated(false);
    setToken(null);
    navigate("/");
  };

  return (
    <AuthContext.Provider
      value={{ login, logout, loading, isAuthenticated, error, token }}
    >
      {children}
    </AuthContext.Provider>
  );
};
