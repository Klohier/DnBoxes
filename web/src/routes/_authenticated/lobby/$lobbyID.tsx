import { createFileRoute } from "@tanstack/react-router";
import { useParams, useNavigate } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Lobby } from "@/types/lobby";
import { useWebSocket } from "@/WebSocketContext";
import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/AuthContext";
import { Message } from "@/types/websocket";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Users, Crown, Check, X, Wifi, WifiOff, LogOut } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

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
  const [isLeaving, setIsLeaving] = useState(false);
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
          event.payload.gameID,
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
        error instanceof Error
          ? error.message
          : "Failed to update ready status",
      );
    } finally {
      setIsTogglingReady(false);
    }
  };

  const handleLeaveLobby = async () => {
    if (isLeaving) return;

    setIsLeaving(true);

    try {
      const res = await fetch(`/api/v1/lobbies/${lobbyID}/leave`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!res.ok) {
        const error = await res.json();
        throw new Error(error.error || "Failed to leave lobby");
      }

      // Navigate back to home/lobby list
      void navigate({ to: "/" });
    } catch (error) {
      console.error("Error leaving lobby:", error);
      alert(error instanceof Error ? error.message : "Failed to leave lobby");
    } finally {
      setIsLeaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-gray-400 text-lg">Loading lobby...</p>
      </div>
    );
  }

  const currentUserID = user?.userID;

  // Check if current user is ready
  const currentPlayer = lobby?.players?.find(
    (p) => p.user_id === currentUserID,
  );
  const isCurrentUserReady = currentPlayer?.is_ready ?? false;

  // Check if all players are ready
  const allPlayersReady = lobby?.players?.every((p) => p.is_ready) ?? false;
  const hasEnoughPlayers = (lobby?.players?.length ?? 0) >= 2;

  // Check if current user is host
  const isHost = lobby?.host_id === currentUserID;

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-3xl space-y-6">
        {/* Header Card */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-2xl text-white">
                  {lobby?.name}
                </CardTitle>
                <CardDescription className="text-gray-400">
                  Waiting for players to ready up
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <Badge
                  variant={connected ? "default" : "destructive"}
                  className="flex items-center gap-1"
                >
                  {connected ? (
                    <>
                      <Wifi className="h-3 w-3" />
                      Connected
                    </>
                  ) : (
                    <>
                      <WifiOff className="h-3 w-3" />
                      Disconnected
                    </>
                  )}
                </Badge>
                <Button
                  onClick={handleLeaveLobby}
                  disabled={isLeaving}
                  variant="destructive"
                  size="sm"
                >
                  <LogOut className="h-4 w-4 mr-2" />
                  {isLeaving ? "Leaving..." : "Leave"}
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div className="flex items-center gap-2 text-gray-400">
                <Users className="h-4 w-4" />
                <span>
                  Players: {lobby?.players?.length ?? 0} / {lobby?.player_limit}
                </span>
              </div>
              <div className="flex items-center gap-2 text-gray-400">
                <span>
                  Board Size: {lobby?.board_size} × {lobby?.board_size}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Players Card */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <CardTitle className="text-xl text-white">Players</CardTitle>
          </CardHeader>
          <CardContent>
            {lobby?.players && lobby.players.length > 0 ? (
              <div className="space-y-2">
                {lobby.players.map((player) => (
                  <div
                    key={player.user_id}
                    className="flex items-center justify-between p-4 bg-gray-700/50 rounded-lg hover:bg-gray-700 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold">
                        {player.username.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-white font-medium">
                            {player.username}
                          </span>
                          {player.user_id === lobby.host_id && (
                            <Crown className="h-4 w-4 text-yellow-500" />
                          )}
                          {player.user_id === currentUserID && (
                            <Badge variant="outline" className="text-xs">
                              You
                            </Badge>
                          )}
                        </div>
                      </div>
                    </div>
                    <Badge
                      variant={player.is_ready ? "default" : "secondary"}
                      className={
                        player.is_ready
                          ? "bg-green-600 hover:bg-green-700"
                          : "bg-gray-600 hover:bg-gray-700"
                      }
                    >
                      {player.is_ready ? (
                        <Check className="h-3 w-3 mr-1" />
                      ) : (
                        <X className="h-3 w-3 mr-1" />
                      )}
                      {player.is_ready ? "Ready" : "Not Ready"}
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-400 text-center py-8">
                No players in lobby
              </p>
            )}
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <Card className="bg-gray-800 border-gray-700">
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="flex gap-3">
                <Button
                  onClick={handleToggleReady}
                  disabled={isTogglingReady || !connected}
                  variant={isCurrentUserReady ? "secondary" : "default"}
                  className="flex-1"
                  size="lg"
                >
                  {isTogglingReady
                    ? "Updating..."
                    : isCurrentUserReady
                      ? "Unready"
                      : "Ready Up"}
                </Button>

                {isHost && (
                  <Button
                    onClick={handleStartGame}
                    disabled={
                      !hasEnoughPlayers ||
                      !allPlayersReady ||
                      isStarting ||
                      !connected
                    }
                    className="flex-1"
                    size="lg"
                  >
                    {isStarting ? "Creating Game..." : "Start Game"}
                  </Button>
                )}
              </div>

              {/* Status Messages */}
              <div className="space-y-2">
                {!allPlayersReady && hasEnoughPlayers && (
                  <div className="flex items-center gap-2 text-yellow-500 text-sm bg-yellow-500/10 p-3 rounded-lg">
                    <span className="text-lg">⚠️</span>
                    <span>All players must be ready before starting</span>
                  </div>
                )}
                {!hasEnoughPlayers && (
                  <div className="flex items-center gap-2 text-red-400 text-sm bg-red-500/10 p-3 rounded-lg">
                    <span className="text-lg">🚫</span>
                    <span>Need at least 2 players to start</span>
                  </div>
                )}
                {!connected && (
                  <div className="flex items-center gap-2 text-red-400 text-sm bg-red-500/10 p-3 rounded-lg">
                    <WifiOff className="h-4 w-4" />
                    <span>WebSocket disconnected - reconnecting...</span>
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
