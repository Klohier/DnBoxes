import {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
  ReactNode,
} from "react";
import { useAuth } from "./AuthContext";

interface WebSocketContextValue {
  socket: WebSocket | null;
  subscribe: (callback: (message: any) => void) => () => void;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

interface WebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
}) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const { isAuthenticated } = useAuth();
  const subscribers = useRef<Array<(message: any) => void>>([]);
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";

  //TODO: Turn this into a custom hook
  useEffect(() => {
    if (!isAuthenticated) {
      if (socket) {
        socket.close();
        setSocket(null);
      }
      return;
    }
    const ws = new WebSocket(`ws://${apiUrl}/api/v1/ws`);
    let reconnectInterval: NodeJS.Timeout;
    const connect = () => {
      ws.onopen = () => {
        console.log("WebSocket connected");
        setSocket(ws); // set only once connected
      };
      ws.onmessage = (event) => {
        console.log("Message received:", event.data);
        const message = JSON.parse(event.data);
        subscribers.current.forEach((cb) => cb(message));
      };
      ws.onclose = () => {
        console.log("WebSocket disconnected");
        reconnectInterval = setTimeout(connect, 1000);
      };
      ws.onerror = (err) => {
        console.error("WebSocket error", err);
        ws.close(); // Trigger onclose
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectInterval);
      ws?.close();
      setSocket(null);
    };
  }, [isAuthenticated]);

  const subscribe = (callback: (message: any) => void) => {
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

export const useWebSocket = (): WebSocketContextValue => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};
