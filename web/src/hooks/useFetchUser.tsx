import { useState, useEffect } from "react";
import axios from "axios";
import { useAuth } from "../AuthContext";

const useFetchUser = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { isAuthenticated, token } = useAuth();
  const apiUrl = (import.meta.env.VITE_API_URL as string) || "localhost:8484";

  useEffect(() => {
    console.log("useEffect triggered:", { isAuthenticated, token });
    if (isAuthenticated) {
      // const decodedToken = atob(token);
      // const tokenParts = decodedToken.split("|");

      // const userId = String(tokenParts[0]);

      // console.log("Fetching user with ID:", userId);
      axios
        .get(`http://${apiUrl}/api/v1/users/me`, {
          withCredentials: true,
        })
        .then((response) => {
          setUser(response.data);
          setLoading(false);
        })
        .catch((error) => {
          console.error("Error fetching user data:", error);
          setLoading(false);
          setError(error);
        });
    } else {
      setLoading(false);
      setUser(null);
    }
  }, [isAuthenticated]);
  return { user, loading, error };
};

export default useFetchUser;
