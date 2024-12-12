import { createContext, useContext, useEffect, useState } from "react";

const WebSocketContext = createContext(null);

// eslint-disable-next-line react/prop-types
export const WebSocketProvider = ({ children }) => {
  const [socket, setSocket] = useState(null);

  useEffect(() => {
    // Create WebSocket connection
    const ws = new WebSocket("ws://localhost:8484/api/v1/ws");
    setSocket(ws);

    // Handle WebSocket events
    ws.onopen = () => console.log("WebSocket connected");
    ws.onmessage = (event) => console.log("Message received:", event.data);
    ws.onclose = () => console.log("WebSocket disconnected");

    return () => {
      // Clean up WebSocket on unmount
      ws.close();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={socket}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  return useContext(WebSocketContext);
};
