import { createContext, useState, useContext, useEffect } from "react";
import axios from "axios";
import Cookies from "js-cookie";
import { useNavigate } from "react-router";

// Create a Context for the user data
const UserContext = createContext();

// Custom hook to use the UserContext
export const useUser = () => {
  return useContext(UserContext);
};

// Provider component to wrap your application with user state
// eslint-disable-next-line react/prop-types
export const UserProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const token = Cookies.get("DnB-Session"); // Retrieve userId from cookie

    if (token) {
      const decodedToken = atob(token);
      const tokenParts = decodedToken.split("|");

      const userId = String(tokenParts[0]);

      console.log(userId);

      // If userId exists, make an API call to fetch user data
      axios
        .get(`http://localhost:8484/api/v1/users/${userId}`) // Replace with your API endpoint
        .then((response) => {
          setUser(response.data); // Store the user data in state
        })
        .catch((error) => {
          console.error("Error fetching user data:", error);
        });
    } else {
      // navigate("/");
    }
  }, [navigate]);

  const loginUser = (userData) => {
    setUser(userData);
  };

  const logoutUser = () => {
    setUser(null);
  };

  return (
    <UserContext.Provider value={{ user, loginUser, logoutUser }}>
      {children}
    </UserContext.Provider>
  );
};
