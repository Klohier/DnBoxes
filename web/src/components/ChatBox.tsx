import { useState, useEffect, useRef, KeyboardEvent } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

interface ChatMessage {
  userID: number;
  username: string;
  session_id: number;
  message: string;
  timestamp: string;
}

interface ChatboxProps {
  sessionID: number | null;
}

const Chatbox: React.FC<ChatboxProps> = ({ sessionID }) => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [newMessage, setNewMessage] = useState<string>("");
  const { user } = useUser();
  const { socket, subscribe } = useWebSocket();
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!sessionID || !socket || socket.readyState !== WebSocket.OPEN) return;

    const fetchMessages = async () => {
      try {
        const endpoint = `http://${apiUrl}/api/v1/chat?sessionID=${sessionID}`;
        const response = await axios.get<ChatMessage[]>(endpoint);
        if (response.data) setMessages(response.data);
      } catch (error) {
        console.error("Error fetching past messages:", error);
      }
    };

    fetchMessages();
  }, [sessionID, socket, apiUrl]);

  useEffect(() => {
    if (!socket || socket.readyState !== WebSocket.OPEN) return;

    const unsubscribe = subscribe((message: any) => {
      if (
        message.type === "chat:new" &&
        message.payload.session_id === sessionID
      ) {
        setMessages((prev) => [...prev, message.payload]);
      }
    });

    return () => unsubscribe();
  }, [sessionID, socket, subscribe]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSendMessage = () => {
    if (newMessage.trim() === "") return;

    if (!user || !sessionID) {
      console.warn("User or session ID missing");
      return;
    }

    const message = {
      type: "chat:new",
      payload: {
        userID: Number(user.userID),
        username: user.username,
        session_id: sessionID,
        message: newMessage,
        timestamp: new Date().toISOString(),
      },
    };

    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(message));
    } else {
      console.log("WebSocket is not open");
    }

    setNewMessage("");
  };

  const onInputKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="flex flex-col border rounded-md w-full max-w-full h-[400px] bg-muted/30 backdrop-blur-sm">
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto px-4 py-2 text-sm space-y-1 font-mono"
      >
        {messages.map((msg, index) => {
          const isOwn = msg.userID === Number(user?.userID);
          return (
            <div key={index} className="text-foreground">
              <span className="text-muted-foreground mr-1 text-[10px]">
                [
                {new Date(msg.timestamp).toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
                ]
              </span>
              <span
                className={`font-bold mr-2 ${
                  isOwn ? "text-blue-400" : "text-yellow-400"
                }`}
              >
                {msg.username}
              </span>
              <span>{msg.message}</span>
            </div>
          );
        })}
      </div>

      <div className="flex p-2 border-t bg-background">
        <Input
          placeholder="Press Enter to send..."
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          onKeyDown={onInputKeyDown}
          className="flex-1 mr-2 text-base"
        />
        <Button
          onClick={handleSendMessage}
          disabled={!newMessage.trim()}
          variant="secondary"
        >
          Send
        </Button>
      </div>
    </div>
  );
};

export default Chatbox;
