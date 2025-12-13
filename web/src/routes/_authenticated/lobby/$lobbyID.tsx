import { createFileRoute } from "@tanstack/react-router";
import { useParams, useNavigate } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Lobby } from "@/types/lobby";
import { useWebSocket } from "@/WebSocketContext";
import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/AuthContext";
import { Message } from "@/types/websocket";

export const Route = createFileRoute("/_authenticated/lobby/$lobbyID")({
  component: LobbyPage,
});

function LobbyPage() {
  const { lobbyID } = useParams({ from: "/_authenticated/lobby/$lobbyID" });
  const { send, subscribe, connected } = useWebSocket();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [isStarting, setIsStarting] = useState(false);
  const [isTogglingReady, setIsTogglingReady] = useState(false);
  const { user } = useAuth();

  const { data: lobby, isLoading } = useQuery<Lobby>({
    queryKey: ["lobby", lobbyID],
    queryFn: async () => {
      const res = await fetch(`/api/v1/lobbies/${lobbyID}`);
      if (!res.ok) throw new Error("Failed to fetch lobby");
      return res.json();
    },
    enabled: !!lobbyID,
  });

  useEffect(() => {
    if (!lobbyID) return;

    console.log("Setting up WebSocket listener for lobby:", lobbyID);

    const unsubscribe = subscribe((event: Message) => {
      console.log("WebSocket event received:", event);

      // Only process events for this specific lobby
      if (event.topic !== `lobby:${lobbyID}`) {
        console.log("Ignoring event - wrong topic:", event.topic);
        return;
      }

      // Handle lobby_updated event
      if (event.type === "lobby_updated") {
        console.log("Updating lobby with payload:", event.payload);

        queryClient.setQueryData<Lobby>(["lobby", lobbyID], (old) => {
          if (!old) {
            console.log("No cached lobby data, skipping update");
            return old;
          }

          // Merge the players update
          return {
            ...old,
            players: event.payload.players ?? old.players,
          };
        });
      }

      // Handle game_started event - navigate all players to the game
      if (event.type === "game:new") {
        console.log(
          "Game started, navigating to game page:",
          event.payload.gameID
        );
        void navigate({ to: `/game/${event.payload.gameID}` });
      } else {
        console.log("Ignoring event - unhandled type:", event.type);
      }
    });

    return () => {
      console.log("Cleaning up WebSocket listener for lobby:", lobbyID);
      unsubscribe();
    };
  }, [lobbyID, subscribe, queryClient, navigate]);

  const handleStartGame = async () => {
    if (!lobby || isStarting) return;

    setIsStarting(true);

    try {
      // Extract player IDs from the lobby
      const playerIds = lobby.players?.map((p) => p.user_id) ?? [];

      // Create game
      const res = await fetch(`/api/v1/games`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          player_ids: playerIds,
          board_size: lobby.board_size,
        }),
      });

      if (!res.ok) {
        const error = await res.json();
        throw new Error(error.error || "Failed to create game");
      }

      const game = await res.json();
      console.log("Game created successfully:", game);

      // Navigate to the game page using the game_id from response
      //TODO: Fix inconsistent api naming
      void navigate({ to: `/game/${game.game_id}` });

      //Send WebSocket message to notify other players
      if (connected) {
        send({
          topic: `lobby:${lobbyID}`,
          type: "game:new",
          payload: { gameID: game.game_id },
        });
      }
    } catch (error) {
      console.error("Error creating game:", error);
      alert(error instanceof Error ? error.message : "Failed to create game");
      setIsStarting(false);
    }
  };

  const handleToggleReady = async () => {
    if (isTogglingReady || !connected) return;

    setIsTogglingReady(true);

    try {
      const res = await fetch(`/api/v1/lobbies/${lobbyID}/ready`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!res.ok) {
        const error = await res.json();
        throw new Error(error.error || "Failed to toggle ready status");
      }

      // The WebSocket will handle updating the UI via lobby_updated event
    } catch (error) {
      console.error("Error toggling ready status:", error);
      alert(
        error instanceof Error ? error.message : "Failed to update ready status"
      );
    } finally {
      setIsTogglingReady(false);
    }
  };

  if (isLoading) return <p>Loading lobby...</p>;

  const currentUserID = user?.userID;

  // Check if current user is ready
  const currentPlayer = lobby?.players?.find(
    (p) => p.user_id === currentUserID
  );
  const isCurrentUserReady = currentPlayer?.is_ready ?? false;

  // Check if all players are ready
  const allPlayersReady = lobby?.players?.every((p) => p.is_ready) ?? false;
  const hasEnoughPlayers = (lobby?.players?.length ?? 0) >= 2;

  // Check if current user is host
  const isHost = lobby?.host_id === currentUserID;

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      <h1>{lobby?.name}</h1>

      <div style={{ marginBottom: "20px" }}>
        <p>Host ID: {lobby?.host_id}</p>
        <p>
          Players: {lobby?.players?.length ?? 0} / {lobby?.player_limit}
        </p>
        <p>
          Board Size: {lobby?.board_size} x {lobby?.board_size}
        </p>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h2>Players in Lobby</h2>
        {lobby?.players && lobby.players.length > 0 ? (
          <ul style={{ listStyle: "none", padding: 0 }}>
            {lobby.players.map((player) => (
              <li
                key={player.user_id}
                style={{
                  padding: "10px",
                  marginBottom: "5px",
                  backgroundColor: "#f5f5f5",
                  borderRadius: "5px",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <span>
                  <strong>{player.username}</strong>
                  {player.user_id === lobby.host_id && " 👑"}
                </span>
                <span
                  style={{
                    color: player.is_ready ? "green" : "orange",
                    fontWeight: "bold",
                  }}
                >
                  {player.is_ready ? "✓ Ready" : "○ Not Ready"}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p>No players in lobby</p>
        )}
      </div>

      <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
        <button
          onClick={handleToggleReady}
          disabled={isTogglingReady || !connected}
          style={{
            padding: "10px 20px",
            fontSize: "16px",
            cursor: isTogglingReady || !connected ? "not-allowed" : "pointer",
            opacity: isTogglingReady || !connected ? 0.5 : 1,
            backgroundColor: "#4CAF50",
            color: "white",
            border: "none",
            borderRadius: "5px",
          }}
        >
          {isTogglingReady
            ? "Updating..."
            : isCurrentUserReady
              ? "Unready"
              : "Ready Up"}
        </button>

        {isHost && (
          <button
            onClick={handleStartGame}
            disabled={
              !hasEnoughPlayers || !allPlayersReady || isStarting || !connected
            }
            style={{
              padding: "10px 20px",
              fontSize: "16px",
              cursor:
                !hasEnoughPlayers ||
                !allPlayersReady ||
                isStarting ||
                !connected
                  ? "not-allowed"
                  : "pointer",
              opacity:
                !hasEnoughPlayers ||
                !allPlayersReady ||
                isStarting ||
                !connected
                  ? 0.5
                  : 1,
              backgroundColor: "#2196F3",
              color: "white",
              border: "none",
              borderRadius: "5px",
            }}
          >
            {isStarting ? "Creating Game..." : "Start Game"}
          </button>
        )}
      </div>

      <div style={{ marginTop: "10px" }}>
        {!allPlayersReady && hasEnoughPlayers && (
          <p style={{ color: "orange", margin: "5px 0" }}>
            All players must be ready before starting
          </p>
        )}
        {!hasEnoughPlayers && (
          <p style={{ color: "red", margin: "5px 0" }}>
            Need at least 2 players to start
          </p>
        )}
        {!connected && (
          <p style={{ color: "red", margin: "5px 0" }}>
            WebSocket disconnected - reconnecting...
          </p>
        )}
      </div>
    </div>
  );
}
