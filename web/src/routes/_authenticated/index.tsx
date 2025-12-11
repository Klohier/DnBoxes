import { useEffect, useState } from "react";
import { useWebSocket } from "@/WebSocketContext";
import { LobbyList } from "@/components/Lobby";
import { LobbyModal } from "@/components/LobbyModal";
import Chatbox from "../../components/ChatBox";
import { Button } from "@/components/ui/button";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Lobby } from "../../types/lobby";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  const navigate = useNavigate();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const queryClient = useQueryClient();
  const { subscribe, send } = useWebSocket();
  // Fetch lobbies
  const { data: lobbies = [], isLoading } = useQuery<Lobby[]>({
    queryKey: ["lobbies"],
    queryFn: async () => {
      const res = await fetch("/api/v1/lobbies");
      if (!res.ok) throw new Error("Failed to fetch lobbies");
      return res.json();
    },
  });

  // Handle creating a lobby
  const handleCreateLobby = async (values: Partial<Lobby>) => {
    const res = await fetch("/api/v1/lobbies", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(values),
    });
    if (!res.ok) {
      console.error("Failed to create lobby");
      return;
    }
    const newLobby = await res.json();
    // Update the query cache so LobbyList updates automatically
    queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) => [
      ...old,
      newLobby,
    ]);
    setIsModalOpen(false);
    navigate({
      to: "/lobby/$lobbyID",
      params: { lobbyID: newLobby.lobby_id },
    });
  };

  useEffect(() => {
    const unsubscribe = subscribe((msg) => {
      if (msg.type === "lobby_created") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) => [
          ...old,
          msg.payload,
        ]);
      } else if (msg.type === "lobby_deleted") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
          old.filter((l) => l.lobby_id !== msg.payload.lobby_id)
        );
      } else if (msg.type === "lobby_updated") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
          old.map((l) =>
            l.lobby_id === msg.payload.lobby_id ? { ...l, ...msg.payload } : l
          )
        );
      }
    });

    return unsubscribe;
  }, [subscribe, queryClient]);

  return (
    <>
      <Button onClick={() => setIsModalOpen(true)}>+ Create Lobby</Button>
      <Chatbox sessionID={1} />

      <LobbyModal
        open={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateLobby}
      />

      {isLoading ? <p>Loading lobbies...</p> : <LobbyList lobbies={lobbies} />}
    </>
  );
}
