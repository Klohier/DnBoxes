import { useEffect, useState } from "react";
import { useWebSocket } from "../WebSocketContext";

export interface Lobby {
  lobby_id: string;
  name: string;
  host_id: number;
  player_limit: number;
  is_private: boolean;
  created_at: string;
}

const LobbyList: React.FC = () => {
  const { subscribe, send, connected } = useWebSocket();
  const [lobbies, setLobbies] = useState<Lobby[]>([]);

  const createLobby = async () => {
    const response = await fetch("/api/v1/lobbies", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
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

  const fetchLobbies = async () => {
    try {
      const response = await fetch("/api/v1/lobbies");
      if (!response.ok) throw new Error("Failed to fetch lobbies");

      const data: Lobby[] = await response.json();
      setLobbies(data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchLobbies();
    if (!connected) return;

    // Subscribe to global:lobbies topic
    const unsubscribe = subscribe((message) => {
      // Only handle lobby events
      if (message.topic !== "global:lobbies") return;

      try {
        const lobbyData = JSON.parse(message.payload as string) as Lobby;
        if (message.type === "lobby_created") {
          setLobbies((prev) => [...prev, lobbyData]);
        } else if (message.type === "lobby_deleted") {
          setLobbies((prev) =>
            prev.filter((lobby) => lobby.lobby_id !== lobbyData.lobby_id)
          );
        }
      } catch (err) {
        console.error("Failed to parse lobby message:", err);
      }
    });

    return () => {
      unsubscribe();
    };
  }, [connected, subscribe, send]);

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
