import { Lobby } from "../types/lobby";
import { Button } from "@/components/ui/button";
import { useNavigate } from "@tanstack/react-router";
import { useQueryClient, useMutation } from "@tanstack/react-query";
import { joinLobby } from "@/api/lobby";

interface LobbyListProps {
  lobbies: Lobby[];
}

export const LobbyList: React.FC<LobbyListProps> = ({ lobbies }) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const joinLobbyMutation = useMutation({
    mutationFn: joinLobby,
    onSuccess: (updatedLobby) => {
      console.log("Joined lobby:", updatedLobby);

      // Update the lobby list with the updated lobby data
      queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
        old.map((l) =>
          l.lobby_id === updatedLobby.lobby_id ? updatedLobby : l,
        ),
      );

      void navigate({
        to: "/lobby/$lobbyID",
        params: { lobbyID: updatedLobby.lobby_id },
      });
    },
    onError: (error) => {
      console.error("Failed to join lobby:", error);
    },
  });

  const handleJoinLobby = (lobbyId: string) => {
    joinLobbyMutation.mutate(lobbyId);
  };

  if (!lobbies.length) return <p>No lobbies currently available.</p>;

  return (
    <ul className="space-y-2">
      {lobbies.map((lobby) => {
        const currentPlayers = lobby.players?.length ?? 0;
        const isJoining = joinLobbyMutation.isPending;

        return (
          <li
            key={lobby.lobby_id}
            className="flex justify-between items-center border p-2 rounded"
          >
            <div>
              <strong>{lobby.name}</strong> | Host: {lobby.host_id} | Players:{" "}
              {currentPlayers}/{lobby.player_limit} |{" "}
              {lobby.is_private ? "Private" : "Public"}
            </div>
            <Button
              size="sm"
              onClick={() => {
                handleJoinLobby(lobby.lobby_id);
              }}
              disabled={isJoining}
            >
              {isJoining ? "Joining..." : "Join Lobby"}
            </Button>
          </li>
        );
      })}
    </ul>
  );
};
