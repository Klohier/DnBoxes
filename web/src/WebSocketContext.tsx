import {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
  ReactNode,
} from "react";
import { useAuth } from "./AuthContext";
import { Message } from "./types/websocket";

interface WebSocketContextValue {
  send: (message: Message) => void;
  subscribe: (callback: (message: any) => void) => () => void;
  connected: boolean;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

interface WebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
}) => {
  const socket = useRef<WebSocket | null>(null);
  const { isAuthenticated, loading } = useAuth();
  const subscribers = useRef<((message: any) => void)[]>([]);
  const [connected, setConnected] = useState(false);
  const apiUrl = (import.meta.env.VITE_API_URL as string) || "localhost:8484";

  //TODO: Turn this into a custom hook
  useEffect(() => {
    if (loading) {
      // Wait until loading finishes, do nothing for now
      return;
    }
    if (!isAuthenticated) {
      if (socket.current) {
        socket.current.close();
        socket.current = null;
      }
      return;
    }
    let reconnectInterval: NodeJS.Timeout;
    const connect = () => {
      if (
        socket.current &&
        (socket.current.readyState === WebSocket.OPEN ||
          socket.current.readyState === WebSocket.CONNECTING)
      ) {
        console.log(
          "WebSocket already connected or connecting. Skipping connect."
        );
        return;
      }
      const ws = new WebSocket(`ws://${apiUrl}/api/v1/ws`);
      socket.current = ws;
      ws.onopen = () => {
        console.log("WebSocket connected");
        setConnected(true);
        // set only once connected
      };
      ws.onmessage = (event) => {
        console.log("Message received:", event.data);
        const message = JSON.parse(event.data);
        subscribers.current.forEach((cb) => {
          cb(message);
        });
      };
      ws.onclose = () => {
        console.log("WebSocket disconnected");
        setConnected(false);
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
      socket.current?.close();
      socket.current = null;
    };
  }, [isAuthenticated, loading]);

  const subscribe = (callback: (message: any) => void) => {
    console.log("Adding subscriber");
    subscribers.current.push(callback);
    return () => {
      console.log("Removing subscriber");
      subscribers.current = subscribers.current.filter((cb) => cb !== callback);
    };
  };
  const send = (message: Message) => {
    if (socket.current && socket.current.readyState === WebSocket.OPEN) {
      socket.current.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket not ready");
    }
  };

  return (
    <WebSocketContext.Provider value={{ send, subscribe, connected }}>
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
