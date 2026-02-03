import { fetchGame } from "@/api/fetchGame";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useEffect, useState, useMemo, useRef, useCallback } from "react";
import { Toaster, toast } from "sonner";
import Grid from "../../components/Grid";
import { useWebSocket } from "@/WebSocketContext";
import { useSound } from "@/hooks/useSound";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { GamePlayer, Game, Message, TimerStatePayload } from "@/types/websocket";
import Chatbox from "@/components/ChatBox";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/AuthContext";
import { Trophy, Users, LogOut, Wifi, WifiOff, Clock, Timer, AlertTriangle } from "lucide-react";

export const gameDetailQuery = (id: string) => ({
  queryKey: ["game", id],
  queryFn: () => fetchGame(id),
});

interface LoaderData {
  gameState: Game | undefined;
}

export const Route = createFileRoute("/game/$gameID")({
  component: RouteComponent,

  loader: async ({ params, context }): Promise<LoaderData> => {
    const query = gameDetailQuery(params.gameID);

    const gameState =
      context.queryClient.getQueryData<Game>(query.queryKey) ??
      (await context.queryClient.fetchQuery(query));

    return {
      gameState: gameState,
    };
  },
});

function RouteComponent() {
  const params = Route.useParams();
  const { gameState: initialGameState } = Route.useLoaderData();
  const { queryClient } = Route.useRouteContext();
  const { user } = useAuth();

  const playEdgeClick = useSound("/click.mp3");
  // const playBotMove = useSound("/click.mp3");

  const { data: gameState, refetch } = useQuery({
    ...gameDetailQuery(params.gameID),
    initialData: initialGameState,
  });

  console.log("gameState", gameState);
  const navigate = useNavigate();
  const { send, subscribe, connected } = useWebSocket();
  const [userColors, setUserColors] = useState<Record<number, string>>({});
  const [isProcessingMove, setIsProcessingMove] = useState(false);
  const [timerState, setTimerState] = useState<TimerStatePayload | null>(null);
  const timerRef = useRef<TimerStatePayload | null>(null);

  const winnerPlayer = gameState?.players.find(
    (p) => p.user_id === gameState.winner_id,
  );

  // Create mapping from turn_order to user_id
  const turnToUserIdMap = useMemo(() => {
    if (!gameState) return {};

    return gameState.players.reduce(
      (acc, player) => {
        acc[player.turn_order] = player.user_id;
        return acc;
      },
      {} as Record<number, number>,
    );
  }, [gameState]);

  useEffect(() => {
    if (!gameState) return;
    console.log("Setting boxes from gameState:", gameState.grid);

    const colors: string[] = [
      "red",
      "blue",
      "green",
      "purple",
      "orange",
      "pink",
    ];
    const colorMap: Record<number, string> = {};
    gameState.players.forEach((player) => {
      colorMap[player.user_id] = colors[player.turn_order % colors.length];
    });
    setUserColors(colorMap);
  }, [gameState]);

  useEffect(() => {
    if (connected && params.gameID) {
      console.log(
        "WebSocket connected, refetching game state to subscribe to room",
      );
      refetch();
    }
  }, [connected, params.gameID, refetch]);

  useEffect(() => {
    console.log("WebSocket connection status:", {
      connected,
      gameID: params.gameID,
      hasGameState: !!gameState,
    });

    if (!params.gameID) {
      console.warn("No gameID available");
      return;
    }

    if (!connected) {
      console.warn("WebSocket not connected yet, waiting...");
      return;
    }

    console.log("Setting up WebSocket listener for game:", params.gameID);

    const unsubscribe = subscribe((message: Message) => {
      console.log("WebSocket message received:", message);

      if (
        message.topic === `game:${params.gameID}` &&
        message.type === "game:state"
      ) {
        console.log("Game state updated via WebSocket:", message.payload);

        // Update the game state in React Query cache
        queryClient.setQueryData(["game", params.gameID], message.payload);

        return;
      }

      // Handle timer sync events from server
      if (
        message.topic === `game:${params.gameID}` &&
        message.type === "game:timer"
      ) {
        const payload = message.payload as TimerStatePayload;
        timerRef.current = payload;
        setTimerState(payload);
        return;
      }
    });

    return () => {
      console.log("Cleaning up WebSocket listener for game:", params.gameID);
      unsubscribe();
    };
  }, [subscribe, params.gameID, connected, queryClient, isProcessingMove]);

  // Format milliseconds as MM:SS
  const formatTime = useCallback((ms: number): string => {
    const totalSeconds = Math.max(0, Math.ceil(ms / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  }, []);

  // Get remaining time for a player from timer state
  const getPlayerTime = useCallback(
    (userId: number): number | null => {
      if (!timerState) return null;
      const playerTimer = timerState.players.find(
        (p) => p.user_id === userId,
      );
      return playerTimer?.remaining_ms ?? null;
    },
    [timerState],
  );

  // Check if a player is disconnected
  const isPlayerDisconnected = useCallback(
    (userId: number): boolean => {
      if (!timerState) return false;
      const playerTimer = timerState.players.find(
        (p) => p.user_id === userId,
      );
      return playerTimer?.disconnected ?? false;
    },
    [timerState],
  );

  // Client-side timer interpolation (smooth countdown between server syncs)
  useEffect(() => {
    if (!timerState || gameState?.winner_id) return;

    const interval = setInterval(() => {
      setTimerState((prev) => {
        if (!prev) return null;
        return {
          ...prev,
          players: prev.players.map((p) => ({
            ...p,
            remaining_ms:
              p.turn_order === prev.active_turn
                ? Math.max(0, p.remaining_ms - 100)
                : p.remaining_ms,
          })),
        };
      });
    }, 100);

    return () => clearInterval(interval);
  }, [timerState?.active_turn, gameState?.winner_id]);

  const handleQuitGame = async () => {
    if (gameState?.game_id && user?.userID && !gameState.winner_id) {
      try {
        await fetch(`/api/v1/games/${gameState.game_id}/forfeit`, {
          method: "POST",
        });
      } catch (error) {
        console.error("Failed to forfeit game:", error);
      }
    }
    navigate({ to: "/" });
  };

  const handleClick = async (
    gameId: number,
    playerId: number,
    row: number,
    col: number,
    edge: string,
  ) => {
    if (isProcessingMove) {
      console.log("Move already in progress, ignoring click");
      return;
    }

    setIsProcessingMove(true);

    try {
      console.log("Making move:", { gameId, playerId, row, col, edge });

      playEdgeClick();

      const response = await fetch(`/api/v1/games/${gameId}/move`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          playerId: playerId,
          row: row,
          col: col,
          edge: edge,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to make move");
      }

      const result = await response.json();
      console.log("Move successful:", result);

      // // Update local state with the returned game
      // if (result) {
      //   queryClient.setQueryData(["game", params.gameID], result);
      // }
    } catch (error) {
      console.error("Failed to make move:", error);
      toast.error(
        error instanceof Error ? error.message : "Failed to make move",
      );
    } finally {
      setIsProcessingMove(false);
    }
  };

  if (!gameState || !user) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <p className="text-lg text-gray-300">Loading game...</p>
          {!connected && (
            <p className="text-sm text-gray-500 mt-2">
              Connecting to server...
            </p>
          )}
        </div>
      </div>
    );
  }

  const currentTurn = gameState.current_turn;
  const currentTurnPlayer = gameState.players.find(
    (p) => p.turn_order === currentTurn,
  );

  const isMyTurn = currentTurnPlayer?.user_id === user.userID;

  const turnDisplayText = isMyTurn
    ? "Your Turn"
    : `${currentTurnPlayer?.username || `Player ${currentTurn}`}'s Turn`;

  // Flatten the 2D grid back to 1D for the Grid component
  const flattenedBoxes = gameState.grid.flat();

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <Toaster position="top-right" richColors />

      {/* Connection Status Banner */}
      {!connected && (
        <div className="fixed top-4 right-4 z-50">
          <Badge
            variant="destructive"
            className="flex items-center gap-2 px-4 py-2"
          >
            <WifiOff className="h-4 w-4" />
            Reconnecting...
          </Badge>
        </div>
      )}

      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* Left/Main Section - Game Board */}
          <div className="lg:col-span-2 space-y-4">
            {/* Game Header */}
            <Card className="bg-gray-800 border-gray-700">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  {/* Turn Indicator */}
                  <div className="flex items-center gap-3">
                    <Clock className="h-5 w-5 text-gray-400" />
                    <div>
                      <p className="text-xs text-gray-400">Current Turn</p>
                      <p
                        className={`text-lg font-semibold ${
                          isMyTurn ? "text-green-400" : "text-white"
                        }`}
                      >
                        {turnDisplayText}
                      </p>
                    </div>
                    {/* Active Player Timer */}
                    {timerState && currentTurnPlayer && (() => {
                      const timeMs = getPlayerTime(currentTurnPlayer.user_id);
                      if (timeMs === null) return null;
                      const isLow = timeMs < 30000;
                      return (
                        <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg font-mono text-lg font-bold ${
                          isLow
                            ? "bg-red-500/20 text-red-400 animate-pulse"
                            : "bg-gray-700 text-white"
                        }`}>
                          <Timer className="h-4 w-4" />
                          {formatTime(timeMs)}
                        </div>
                      );
                    })()}
                  </div>

                  {/* Connection Status & Quit */}
                  <div className="flex items-center gap-2">
                    <Badge
                      variant={connected ? "default" : "destructive"}
                      className="hidden sm:flex items-center gap-1"
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
                      variant="destructive"
                      size="sm"
                      onClick={handleQuitGame}
                    >
                      <LogOut className="h-4 w-4 mr-2" />
                      Quit
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Game Board */}
            <Card className="bg-gray-800 border-gray-700">
              <CardContent className="p-6 flex justify-center">
                <div className="w-full max-w-[650px]">
                  <Grid
                    gameID={gameState.game_id}
                    boxes={flattenedBoxes}
                    userColors={userColors}
                    boardSize={gameState.board_size}
                    userID={user.userID}
                    handleClick={handleClick}
                    turnToUserIdMap={turnToUserIdMap}
                  />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Right Section - Players & Chat */}
          <div className="lg:col-span-1 space-y-4">
            {/* Players/Scoreboard */}
            <Card className="bg-gray-800 border-gray-700">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <Users className="h-5 w-5" />
                  Players
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {gameState.players
                    .sort((a, b) => b.score - a.score)
                    .map((player, index) => (
                      <div
                        key={player.user_id}
                        className={`flex items-center justify-between p-3 rounded-lg transition-colors ${
                          player.user_id === user.userID
                            ? "bg-blue-500/20 border border-blue-500/50"
                            : "bg-gray-700/50"
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          {/* Rank Badge */}
                          {index === 0 && gameState.winner_id && (
                            <Trophy className="h-5 w-5 text-yellow-500" />
                          )}

                          {/* Player Avatar */}
                          <div
                            className="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold text-sm"
                            style={{
                              backgroundColor:
                                userColors[player.user_id] || "#666",
                            }}
                          >
                            {player.username.charAt(0).toUpperCase()}
                          </div>

                          {/* Player Info */}
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="text-white font-medium">
                                {player.user_id === user.userID
                                  ? "You"
                                  : player.username}
                              </span>
                              {currentTurnPlayer?.user_id ===
                                player.user_id && (
                                <Badge variant="outline" className="text-xs">
                                  Active
                                </Badge>
                              )}
                              {isPlayerDisconnected(player.user_id) && (
                                <Badge variant="destructive" className="text-xs flex items-center gap-1">
                                  <AlertTriangle className="h-3 w-3" />
                                  DC
                                </Badge>
                              )}
                            </div>
                            {/* Timer display per player */}
                            {(() => {
                              const timeMs = getPlayerTime(player.user_id);
                              if (timeMs === null) return (
                                <p className="text-xs text-gray-400">
                                  Turn {player.turn_order + 1}
                                </p>
                              );
                              const isLow = timeMs < 30000;
                              const isActive = currentTurnPlayer?.user_id === player.user_id;
                              return (
                                <p className={`text-xs font-mono ${
                                  isLow ? "text-red-400 font-semibold" :
                                  isActive ? "text-green-400" : "text-gray-400"
                                }`}>
                                  {formatTime(timeMs)}
                                </p>
                              );
                            })()}
                          </div>
                        </div>

                        {/* Score */}
                        <div className="text-right">
                          <p className="text-2xl font-bold text-white">
                            {player.score}
                          </p>
                          <p className="text-xs text-gray-400">points</p>
                        </div>
                      </div>
                    ))}
                </div>
              </CardContent>
            </Card>

            {/* Chat */}
            <Card className="bg-gray-800 border-gray-700 overflow-hidden">
              <CardHeader className="pb-3 flex-shrink-0">
                <CardTitle className="text-white text-lg">Game Chat</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {gameState.game_id && (
                  <Chatbox
                    topic={`game:${String(gameState.game_id)}`}
                    gameID={gameState.game_id}
                  />
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      {/* Game Over Dialog */}
      <Dialog open={!!gameState.winner_id}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-center text-2xl">
              Game Over!
            </DialogTitle>
          </DialogHeader>
          <div className="text-center py-6">
            {winnerPlayer && (
              <div className="mb-4">
                <Trophy className="h-16 w-16 text-yellow-500 mx-auto mb-3" />
                <p className="text-xl font-bold text-gray-900">
                  {winnerPlayer.user_id === user.userID
                    ? "You Won!"
                    : `${winnerPlayer.username} Wins!`}
                </p>
              </div>
            )}
            {!winnerPlayer && (
              <p className="text-lg font-semibold text-gray-900">It's a Tie!</p>
            )}
          </div>
          <div className="border-t pt-4">
            <h4 className="font-semibold mb-3 text-gray-900">Final Scores</h4>
            <div className="space-y-2">
              {gameState.players
                .sort((a, b) => b.score - a.score)
                .map((player, index) => (
                  <div
                    key={player.user_id}
                    className="flex items-center justify-between p-2 bg-gray-50 rounded"
                  >
                    <div className="flex items-center gap-2">
                      {index === 0 && (
                        <Trophy className="h-4 w-4 text-yellow-500" />
                      )}
                      <span className="font-medium">
                        {player.user_id === user.userID
                          ? "You"
                          : player.username}
                      </span>
                    </div>
                    <span className="font-bold">{player.score}</span>
                  </div>
                ))}
            </div>
          </div>
          <DialogFooter>
            <Button
              onClick={() => {
                void navigate({ to: "/" });
              }}
              className="w-full"
            >
              Return Home
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
