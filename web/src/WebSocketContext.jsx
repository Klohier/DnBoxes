import { createContext, useContext, useEffect, useState, useRef } from "react";
import { useAuth } from "./AuthContext";
const WebSocketContext = createContext(null);

// eslint-disable-next-line react/prop-types
export const WebSocketProvider = ({ children }) => {
  const [socket, setSocket] = useState(null);
  const { isAuthenticated } = useAuth();
  const subscribers = useRef([]);
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";

  //TODO: Turn this into a custom hook
  useEffect(() => {
    // Create WebSocket connection

    if (!isAuthenticated) {
      // If not authenticated, close existing socket
      if (socket) {
        socket.close();
        setSocket(null);
      }
      return;
    }

    const ws = new WebSocket(`ws://${apiUrl}/api/v1/ws`);

    // Handle WebSocket events
    ws.onopen = () => {
      console.log("WebSocket connected");
      setSocket(ws); // set only once connected
    };
    ws.onmessage = (event) => {
      console.log("Message received:", event.data);
      const message = JSON.parse(event.data);
      subscribers.current.forEach((cb) => cb(message));
    };
    ws.onclose = () => console.log("WebSocket disconnected");

    return () => {
      // Clean up WebSocket on unmount
      ws.close();
      setSocket(null);
    };
  }, [isAuthenticated]);

  const subscribe = (callback) => {
    subscribers.current.push(callback);
    return () => {
      subscribers.current = subscribers.current.filter((cb) => cb !== callback);
    };
  };

  return (
    <WebSocketContext.Provider value={{ socket, subscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  return useContext(WebSocketContext);
};
