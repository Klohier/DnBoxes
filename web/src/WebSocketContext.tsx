import {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
  ReactNode,
  useMemo,
  useCallback,
} from "react";
import { useAuth } from "./AuthContext";
import { Message } from "./types/websocket";

interface WebSocketContextValue {
  send: (message: Message) => void;
  subscribe: (callback: (message: Message) => void) => () => void;
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
  const { isAuthenticated } = useAuth();
  const subscribers = useRef<Set<(message: Message) => void>>(new Set());
  const [connected, setConnected] = useState(false);
  const apiUrl = "localhost";
  const reconnectTimeout = useRef<NodeJS.Timeout | null>(null);

  //TODO: Turn this into a custom hook
  useEffect(() => {
    if (!isAuthenticated) {
      if (socket.current) {
        socket.current.close();
        socket.current = null;
        setConnected(false);
      }
      subscribers.current.clear();
      return;
    }

    if (socket.current) return;

    const connect = () => {
      const ws = new WebSocket(`wss://${apiUrl}/api/v1/ws`);
      socket.current = ws;
      ws.onopen = () => {
        console.log("WebSocket connected");
        setConnected(true);
        if (reconnectTimeout.current) {
          clearTimeout(reconnectTimeout.current);
          reconnectTimeout.current = null;
        }
        // set only once connected
      };
      ws.onmessage = (event: MessageEvent<string>) => {
        console.log("Message received:", event.data);
        const message = JSON.parse(event.data) as Message;
        subscribers.current.forEach((cb) => {
          cb(message);
        });
      };
      ws.onclose = () => {
        console.log("WebSocket disconnected");
        setConnected(false);
        socket.current = null;
        subscribers.current.clear();
        reconnectTimeout.current ??= setTimeout(connect, 1000);
      };
      ws.onerror = (err) => {
        console.error("WebSocket error", err);
        ws.close();
      };
    };

    connect();

    return () => {
      if (socket.current) {
        socket.current.close();
        socket.current = null;
      }
      subscribers.current.clear();
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
        reconnectTimeout.current = null;
      }
    };
  }, [isAuthenticated]);

  const subscribe = useCallback((callback: (message: Message) => void) => {
    console.log("Adding subscriber");
    subscribers.current.add(callback);
    console.log("Total subscribers:", subscribers.current.size);
    return () => {
      console.log("Removing subscriber");
      subscribers.current.delete(callback);
      console.log("Total subscribers:", subscribers.current.size);
    };
  }, []);
  const send = (message: Message) => {
    if (socket.current && socket.current.readyState === WebSocket.OPEN) {
      socket.current.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket not ready");
    }
  };

  const contextValue = useMemo(
    () => ({
      send,
      subscribe,
      connected,
    }),
    [connected]
  );

  return (
    <WebSocketContext.Provider value={contextValue}>
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
