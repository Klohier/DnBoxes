import { createFileRoute } from "@tanstack/react-router";
import { useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Lobby } from "@/types/lobby";
import { useWebSocket } from "@/WebSocketContext";
import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";

export const Route = createFileRoute("/_authenticated/lobby/$lobbyID")({
  component: LobbyPage,
});

function LobbyPage() {
  const { lobbyID } = useParams({ from: "/_authenticated/lobby/$lobbyID" });
  const { send, subscribe, connected } = useWebSocket();
  const queryClient = useQueryClient();

  const { data: lobby, isLoading } = useQuery<Lobby>({
    queryKey: ["lobby", lobbyID],
    queryFn: async () => {
      const res = await fetch(`/api/v1/lobbies/${lobbyID}`);
      if (!res.ok) throw new Error("Failed to fetch lobby");
      return res.json();
    },
    enabled: !!lobbyID,
  });

  // useEffect(() => {
  //   if (!lobbyID) return;

  //   const unsubscribe = subscribe((event: any) => {
  //     if (event.topic !== `lobby:${lobbyID}`) return;
  //     if (event.type === "player_joined" || event.type === "player_left") {
  //       queryClient.setQueryData<Lobby>(["lobby", lobbyID], (old) => {
  //         if (!old) return old;

  //         let updatedPlayers = [...(old.players ?? [])];

  //         if (event.type === "player_joined") {
  //           updatedPlayers.push(event.payload);
  //         } else if (event.type === "player_left") {
  //           updatedPlayers = updatedPlayers.filter(
  //             (p) => p.userID !== event.payload.userID
  //           );
  //         }

  //         return { ...old, players: updatedPlayers };
  //       });
  //     }
  //   });

  //   return () => unsubscribe();
  // }, [lobbyID, subscribe, queryClient]);

  useEffect(() => {
    if (!lobbyID) return;

    console.log("Setting up WebSocket listener for lobby:", lobbyID);

    const unsubscribe = subscribe((event: any) => {
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
            console.log("No old data, using payload as-is");
            return event.payload;
          }

          // Merge the update with existing data
          const updated = { ...old, ...event.payload };
          console.log("Merged lobby data:", updated);
          return updated;
        });
      } else {
        console.log("Ignoring event - unhandled type:", event.type);
      }
    });

    return () => {
      console.log("Cleaning up WebSocket listener for lobby:", lobbyID);
      unsubscribe();
    };
  }, [lobbyID, subscribe, queryClient]);

  if (isLoading) return <p>Loading lobby...</p>;

  return (
    <div>
      <h1>{lobby?.name}</h1>

      <div>
        <p>Host ID: {lobby?.host_id}</p>
        {/* <p>Status: {lobby.status || "waiting"}</p> */}
        <p>
          Players: {lobby?.players?.length || 0} / {lobby?.player_limit}
        </p>
      </div>

      <div>
        <h2>Players in Lobby</h2>
        {lobby?.players && lobby.players.length > 0 ? (
          <ul>
            {lobby.players.map((player) => (
              <li key={player.userID}>
                User {player.userID} -{" "}
                {player.is_ready ? "✓ Ready" : "○ Not Ready"}
              </li>
            ))}
          </ul>
        ) : (
          <p>No players in lobby</p>
        )}
      </div>
    </div>
  );
}
