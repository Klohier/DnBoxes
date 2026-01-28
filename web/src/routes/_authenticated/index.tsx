import { useEffect, useState } from "react";
import { useWebSocket } from "@/WebSocketContext";
import { LobbyList } from "@/components/Lobby";
import { LobbyModal } from "@/components/LobbyModal";
import Chatbox from "../../components/ChatBox";
import { Button } from "@/components/ui/button";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import { CreateLobbyData, Lobby } from "../../types/lobby";
import { fetchLobbies, createLobby } from "@/api/lobby";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  const navigate = useNavigate();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const queryClient = useQueryClient();
  const { subscribe } = useWebSocket();

  // Fetch lobbies
  const { data: lobbies = [], isLoading } = useQuery<Lobby[]>({
    queryKey: ["lobbies"],
    queryFn: fetchLobbies,
  });

  // Create lobby mutation
  const createLobbyMutation = useMutation({
    mutationFn: createLobby,
    onSuccess: (newLobby) => {
      // Update the query cache
      queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) => [
        ...old,
        newLobby,
      ]);
      setIsModalOpen(false);
      void navigate({
        to: "/lobby/$lobbyID",
        params: { lobbyID: newLobby.lobby_id },
      });
    },
    onError: (error) => {
      console.error("Failed to create lobby:", error);
    },
  });

  const handleCreateLobby = async (values: CreateLobbyData): Promise<void> => {
    await createLobbyMutation.mutateAsync(values);
    setIsModalOpen(false);
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
          old.filter((l) => l.lobby_id !== msg.payload.lobby_id),
        );
      } else if (msg.type === "lobby_updated") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
          old.map((l) =>
            l.lobby_id === msg.payload.lobby_id ? { ...l, ...msg.payload } : l,
          ),
        );
      }
    });

    return unsubscribe;
  }, [subscribe, queryClient]);

  return (
    <>
      <Button
        onClick={() => {
          setIsModalOpen(true);
        }}
      >
        + Create Lobby
      </Button>

      <LobbyModal
        open={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
        }}
        onSubmit={handleCreateLobby}
      />

      {isLoading ? <p>Loading lobbies...</p> : <LobbyList lobbies={lobbies} />}
      <Chatbox sessionID={1} />
    </>
  );
}
