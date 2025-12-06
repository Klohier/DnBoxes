import { useEffect } from "react";
import { useWebSocket } from "../WebSocketContext";
import { useQuery, useQueryClient } from "@tanstack/react-query";

export interface Lobby {
  lobby_id: string;
  name: string;
  host_id: number;
  player_limit: number;
  is_private: boolean;
  created_at: string;
  Players?: { is_ready: boolean; userID: number }[];
}

const LobbyList: React.FC = () => {
  const { subscribe, send, connected } = useWebSocket();
  const queryClient = useQueryClient();

  const { data: lobbies = [] } = useQuery<Lobby[]>({
    queryKey: ["lobbies"],
    queryFn: async () => {
      const response = await fetch("/api/v1/lobbies");
      if (!response.ok) throw new Error("Failed to fetch lobbies");
      return response.json();
    },
    enabled: connected,
  });

  const createLobby = async () => {
    const response = await fetch("/api/v1/lobbies", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: "New Lobby",
        player_limit: 4,
        is_private: false,
      }),
    });

    if (!response.ok) {
      console.error("Failed to create lobby");
      return;
    }
  };

  useEffect(() => {
    if (!connected) return;

    const unsubscribe = subscribe((message) => {
      console.log("LobbyList received:", message);

      switch (message.type) {
        case "lobby_created":
          queryClient.setQueryData<Lobby[]>(["lobbies"], (old) => [
            ...(old ?? []),
            message.payload,
          ]);
          break;

        case "lobby_deleted":
          queryClient.setQueryData<Lobby[]>(["lobbies"], (old) =>
            (old ?? []).filter((l) => l.lobby_id !== message.payload.lobby_id)
          );
          break;

        case "lobby_updated":
          queryClient.setQueryData<Lobby[]>(["lobbies"], (old) =>
            (old ?? []).map((l) =>
              l.lobby_id === message.payload.lobby_id
                ? { ...l, ...message.payload }
                : l
            )
          );
          break;
      }
    });

    return () => unsubscribe();
  }, [connected, subscribe, queryClient]);

  return (
    <div>
      <h3>Available Lobbies</h3>

      <button onClick={createLobby}>+ Create Lobby</button>

      {!lobbies.length && <p>No lobbies currently available.</p>}

      <ul>
        {lobbies.map((lobby) => (
          <li key={lobby.lobby_id}>
            <strong>{lobby.name}</strong> | Host: {lobby.host_id} | Players:{" "}
            {lobby.player_limit} | {lobby.is_private ? "Private" : "Public"}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default LobbyList;
