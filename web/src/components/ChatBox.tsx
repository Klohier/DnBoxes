import { useState, useEffect, useRef, KeyboardEvent } from "react";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Message, ChatMessagePayload } from "@/types/websocket";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/AuthContext";

interface ChatboxProps {
  topic: string;
  gameID?: number;
}

const Chatbox: React.FC<ChatboxProps> = ({ topic, gameID }) => {
  const [newMessage, setNewMessage] = useState<string>("");
  const { user } = useAuth();

  const { send, subscribe, connected } = useWebSocket();
  const scrollRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();

  const sendMessageMutation = useMutation({
    mutationFn: (message: Message) => {
      send(message);
      return Promise.resolve(message);
    },
  });

  const { data: fetchedMessages } = useQuery<ChatMessagePayload[]>({
    queryKey: ["chatMessages", topic],
    queryFn: async () => {
      const endpoint =
        gameID != null
          ? `/api/v1/games/${String(gameID)}/chat`
          : `/api/v1/chat`;
      const response = await axios.get<ChatMessagePayload[]>(endpoint);
      return response.data;
    },
  });

  useEffect(() => {
    if (!connected) return;
    const unsubscribe = subscribe((message: Message) => {
      if (message.type === "chat:new" && message.topic === topic) {
        queryClient.setQueryData<ChatMessagePayload[]>(
          ["chatMessages", topic],
          (old) => [...(old ?? []), message.payload],
        );
      }
    });

    return () => {
      unsubscribe();
    };
  }, [topic, subscribe, queryClient, connected]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [fetchedMessages ?? []]);

  const handleSendMessage = () => {
    if (newMessage.trim() === "") return;

    if (!user) {
      console.warn("User not available");
      return;
    }

    const message: Message = {
      type: "chat:new",
      topic: topic,
      payload: {
        userID: user.userID,
        username: user.username,
        message: newMessage,
        timestamp: new Date().toISOString(),
      },
    };

    sendMessageMutation.mutate(message);
    setNewMessage("");
  };

  const onInputKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="flex flex-col h-full w-full">
      {/* Messages area */}
      <div
        ref={scrollRef}
        className="overflow-y-auto px-4 py-2 text-sm space-y-1 font-mono h-[340px]"
      >
        {(fetchedMessages ?? []).map((msg, index) => {
          const isOwn = msg.userID === user?.userID;
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
              <span className="text-white">{msg.message}</span>
            </div>
          );
        })}
      </div>

      {/* Input area  */}
      <div className="flex-shrink-0 p-4 border-t border-gray-700">
        <div className="flex gap-2">
          <Input
            placeholder="Press Enter to send..."
            value={newMessage}
            onChange={(e) => {
              setNewMessage(e.target.value);
            }}
            onKeyDown={onInputKeyDown}
            className="flex-1 text-white"
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
    </div>
  );
};

export default Chatbox;
