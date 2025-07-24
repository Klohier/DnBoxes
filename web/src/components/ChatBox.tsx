import { useState, useEffect, useRef, KeyboardEvent } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Message, ChatMessagePayload } from "@/types/websocket";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

interface ChatboxProps {
  sessionID: number | undefined;
}

const Chatbox: React.FC<ChatboxProps> = ({ sessionID }) => {
  // const [messages, setMessages] = useState<ChatMessagePayload[]>([]);
  const [newMessage, setNewMessage] = useState<string>("");
  const { user } = useUser();

  const { send, subscribe, connected } = useWebSocket();
  const apiUrl = (import.meta.env.VITE_API_URL as string) || "localhost:8484";
  const scrollRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();
  // useEffect(() => {
  //   if (!sessionID) return;

  const sendMessageMutation = useMutation({
    mutationFn: async (message: Message) => {
      // Send via WebSocket here
      send(message);
      return message;
    },

    onError: (_err, _newMessage, context) => {
      // Roll back optimistic update on error
      if (context?.previousMessages) {
        queryClient.setQueryData(
          ["chatMessages", sessionID, apiUrl],
          context.previousMessages
        );
      }
    },
  });

  const {
    data: fetchedMessages,
    isLoading,
    isError,
    error,
  } = useQuery<ChatMessagePayload[]>({
    queryKey: ["chatMessages", sessionID, apiUrl],
    queryFn: async () => {
      if (!sessionID) return [];
      const endpoint = `http://${apiUrl}/api/v1/chat?sessionID=${sessionID}`;
      const response = await axios.get<ChatMessagePayload[]>(endpoint);
      return response.data;
    },
    enabled: !!sessionID,
    staleTime: 1000 * 60,
  });

  useEffect(() => {
    console.log("Effect triggered", { connected, sessionID });
    if (!sessionID || !connected) return;
    const unsubscribe = subscribe((message: Message) => {
      console.log("Received WS message:", message);
      if (
        message.type === "chat:new" &&
        message.payload.session_id === sessionID
      ) {
        queryClient.setQueryData<ChatMessagePayload[]>(
          ["chatMessages", sessionID, apiUrl],
          (old) => [...(old ?? []), message.payload]
        );
      }
    });

    return () => {
      unsubscribe();
    };
  }, [sessionID, subscribe, queryClient, apiUrl, connected]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [fetchedMessages ?? []]);

  const handleSendMessage = () => {
    if (newMessage.trim() === "") return;

    if (!user || !sessionID) {
      console.warn("User or session ID missing");
      return;
    }

    const message: Message = {
      type: "chat:new",
      payload: {
        userID: Number(user.userID),
        username: user.username,
        session_id: sessionID,
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
    <div className="flex flex-col border rounded-md w-full max-w-full h-[400px] bg-muted/30 backdrop-blur-sm">
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto px-4 py-2 text-sm space-y-1 font-mono"
      >
        {(fetchedMessages ?? []).map((msg, index) => {
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
          onChange={(e) => {
            setNewMessage(e.target.value);
          }}
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
