import { fetchGame } from "@/api/fetchGame";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useEffect, useState, useMemo } from "react";
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
import { GamePlayer, Game, Message } from "@/types/websocket";
import Chatbox from "@/components/ChatBox";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/AuthContext";

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

  const winnerPlayer = gameState?.players.find(
    (p) => p.user_id === gameState.winner_id
  );

  // Create mapping from turn_order to user_id
  const turnToUserIdMap = useMemo(() => {
    if (!gameState) return {};

    return gameState.players.reduce(
      (acc, player) => {
        acc[player.turn_order] = player.user_id;
        return acc;
      },
      {} as Record<number, number>
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
    gameState.players.forEach((player, index) => {
      colorMap[player.user_id] = colors[index % colors.length];
    });
    setUserColors(colorMap);
  }, [gameState]);

  useEffect(() => {
    if (connected && params.gameID) {
      console.log(
        "WebSocket connected, refetching game state to subscribe to room"
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

        // if (!isProcessingMove) {
        //   playBotMove();
        // }

        // Update the game state in React Query cache
        queryClient.setQueryData(["game", params.gameID], message.payload);

        return;
      }
    });

    return () => {
      console.log("Cleaning up WebSocket listener for game:", params.gameID);
      unsubscribe();
    };
  }, [subscribe, params.gameID, connected, queryClient, isProcessingMove]);

  const handleQuitGame = () => {
    if (gameState?.game_id && user?.userID) {
      send({
        type: "game:quit",
        payload: {
          gameId: gameState.game_id,
          playerId: user.userID,
        },
      });
    }
    navigate({ to: "/" });
  };

  const handleClick = async (
    gameId: number,
    playerId: number,
    row: number,
    col: number,
    edge: string
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
        error instanceof Error ? error.message : "Failed to make move"
      );
    } finally {
      setIsProcessingMove(false);
    }
  };

  if (!gameState || !user) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <p className="text-lg">Loading game...</p>
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
    (p) => p.turn_order === currentTurn
  );

  const isMyTurn = currentTurnPlayer?.user_id === user.userID;

  const turnDisplayText = isMyTurn
    ? "Your Turn"
    : `${currentTurnPlayer?.username || `Player ${currentTurn}`}'s Turn`;

  // Flatten the 2D grid back to 1D for the Grid component
  const flattenedBoxes = gameState.grid.flat();

  return (
    <div className="flex flex-col md:flex-row h-screen p-4 gap-4">
      <Toaster position="top-right" richColors />

      {!connected && (
        <div className="fixed top-4 right-4 bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-2 rounded">
          Reconnecting...
        </div>
      )}

      <div className="w-full md:w-2/3 flex flex-col items-center">
        <div className="flex justify-between items-center w-full max-w-[700px] px-4 mb-2">
          <p
            className={`text-lg font-medium ${
              isMyTurn ? "text-green-600 font-bold" : ""
            }`}
          >
            Current Turn: {turnDisplayText}
          </p>
          <div className="mt-2">
            <h3 className="font-semibold mb-1">Scores:</h3>
            <ul className="text-sm space-y-1">
              {gameState.players.map((player) => (
                <li
                  key={player.user_id}
                  className={player.user_id === user.userID ? "font-bold" : ""}
                >
                  {player.user_id === user.userID ? "You" : player.username}:{" "}
                  {player.score}
                </li>
              ))}
            </ul>
          </div>
          <Button variant="destructive" onClick={handleQuitGame}>
            Quit Game
          </Button>
        </div>
        <div className="w-full max-w-[500px] md:max-w-[700px] lg:max-w-[650px]">
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
      </div>

      <div className="w-full md:w-1/3 flex flex-col gap-4">
        <div className="flex-1 overflow-y-auto">{/* <PlayerLobby /> */}</div>
        <div className="h-150">
          {gameState.game_id && <Chatbox sessionID={gameState.game_id} />}
        </div>
      </div>

      <Dialog open={!!gameState.winner_id}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Game Over</DialogTitle>
          </DialogHeader>
          <div className="text-center py-4">
            <p className="text-lg font-semibold">
              {winnerPlayer
                ? `Congratulations to ${winnerPlayer.username}!`
                : "It's a tie!"}
            </p>
          </div>
          <div className="mt-4">
            <h4 className="font-semibold mb-2">Final Scores:</h4>
            <ul className="text-sm space-y-1">
              {gameState.players.map((player) => (
                <li key={player.user_id}>
                  {player.username}: {player.score}
                </li>
              ))}
            </ul>
          </div>
          <DialogFooter>
            <Button
              onClick={() => {
                void navigate({ to: "/" });
              }}
            >
              Return Home
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
