import { Lobby } from "../types/lobby";
import { useWebSocket } from "../WebSocketContext";
import { Button } from "@/components/ui/button";
import { useNavigate } from "@tanstack/react-router";
import { useQueryClient } from "@tanstack/react-query";

interface LobbyListProps {
  lobbies: Lobby[];
}

export const LobbyList: React.FC<LobbyListProps> = ({ lobbies }) => {
  const { send } = useWebSocket();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const handleJoinLobby = async (lobbyId: string) => {
    try {
      const res = await fetch(`/api/v1/lobbies/${lobbyId}/join`, {
        method: "POST",
        credentials: "include",
      });

      if (!res.ok) {
        console.error("Failed to join lobby");
        return;
      }

      const updatedLobby: Lobby = await res.json();
      console.log("Joined lobby:", updatedLobby);

      queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
        old.map((l) =>
          l.lobby_id === updatedLobby.lobby_id ? updatedLobby : l
        )
      );

      navigate({
        to: "/lobby/$lobbyID",
        params: { lobbyID: updatedLobby.lobby_id },
      });
    } catch (err) {
      console.error("Error joining lobby:", err);
    }
  };

  if (!lobbies.length) return <p>No lobbies currently available.</p>;

  return (
    <ul className="space-y-2">
      {lobbies.map((lobby) => {
        const currentPlayers = lobby.players?.length ?? 0;

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
            <Button size="sm" onClick={() => handleJoinLobby(lobby.lobby_id)}>
              Join Lobby
            </Button>
          </li>
        );
      })}
    </ul>
  );
};
