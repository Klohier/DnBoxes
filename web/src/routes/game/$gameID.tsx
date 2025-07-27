import { fetchGame } from "@/api/fetchGame";
import {
  createFileRoute,
  useNavigate,
  useRouteContext,
} from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { Toaster, toast } from "sonner";
import Grid from "../../components/Grid";
import { useWebSocket } from "@/WebSocketContext";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Box, GamePlayer, Message } from "@/types/websocket";
import PlayerLobby from "@/components/PlayerLobby";
import Chatbox from "@/components/ChatBox";
import { useSuspenseQuery } from "@tanstack/react-query";

export const Route = createFileRoute("/game/$gameID")({
  component: RouteComponent,

  loader: ({ params, context }) => {
    const gameState = context.queryClient.ensureQueryData({
      queryKey: ["game", params.gameID],
      queryFn: () => fetchGame(params.gameID),
    });

    // Return gameID to the component (data is cached)
    return {
      gameState: gameState,
      user: context.authentication.user,
    };
  },
});

function RouteComponent() {
  const params = Route.useParams();

  const { user } = Route.useLoaderData();
  const { queryClient } = useRouteContext({ from: "/game/$gameID" });

  const { data: gameState } = useSuspenseQuery({
    queryKey: ["game", params.gameID],
    queryFn: () => fetchGame(params.gameID),
  });
  console.log("gameState", gameState);
  const navigate = useNavigate();
  const [winnerUsername, setWinnerUsername] = useState<string | null>(null);
  const { send, subscribe, connected } = useWebSocket();
  const [userColors, setUserColors] = useState<Record<string, string>>({});
  const [winnerId, setWinnerId] = useState<number | null>(null);

  const winnerPlayer = gameState?.game?.players.find(
    (p) => p.user_id === gameState?.game?.winner
  );

  useEffect(() => {
    if (!gameState?.game) return;
    console.log("Setting boxes from gameState:", gameState.grids);

    const colors: string[] = [
      "red",
      "blue",
      "green",
      "purple",
      "orange",
      "pink",
    ];
    const colorMap: Record<GamePlayer["user_id"], string> = {};
    gameState.game.players.forEach((player, index) => {
      colorMap[player.user_id] = colors[index % colors.length];
    });
    setUserColors(colorMap);
  }, [gameState]);

  useEffect(() => {
    if (!connected) return;
    const colors: string[] = [
      "red",
      "blue",
      "green",
      "purple",
      "orange",
      "pink",
    ];

    const unsubscribe = subscribe((message: Message) => {
      if (message.type === "game:state") {
        queryClient.setQueryData(["game", params.gameID], message.payload);
        if (gameState?.game?.players) {
          const colorMap: Record<GamePlayer["user_id"], string> = {};
          gameState.game.players.forEach((player, index) => {
            colorMap[player.user_id] = colors[index % colors.length];
          });
          setUserColors(colorMap);
        }
      }

      if (message.type === "winner_set") {
        const winnerId = message.payload.winnerId;
        const winnerUsername = message.payload.winnerUsername;
        setWinnerId(winnerId);
        setWinnerUsername(winnerUsername);
        toast.success(`Player ${winnerUsername} has won the game!`);
      }

      if (message.type === "your_turn") {
        toast.info("It's your turn!", {
          description: "Make your move now.",
        });
      }

      if (message.type === "invalid_move") {
        toast.error("Invalid Move: Not your turn or already selected.");
      }

      if (message.type === "game:quit") {
        console.log("A user quit the game");
      }
    });

    return () => {
      unsubscribe();
    };
  }, [subscribe, gameState, connected]);

  const handleQuitGame = () => {
    if (gameState?.game?.session_id && user?.userID) {
      send({
        type: "game:quit",
        payload: {
          gameId: gameState.game.game_id,
          playerId: user.userID,
          session_id: gameState.game.session_id,
        },
      });
    }
    // navigate("/home");
  };

  const handleClick = (
    gameId: number,
    playerId: number,
    row: number,
    col: number,
    edge: string
  ) => {
    const payload = {
      gameId,
      playerId,
      row,
      col,
      edge,
    };

    console.log(payload);

    send({
      type: "game:move",
      payload,
    });
  };

  return (
    <div className="flex flex-col md:flex-row h-screen p-4 gap-4">
      <Toaster position="top-right" richColors />

      {/* Grid section */}
      <div className="w-full md:w-2/3 flex flex-col items-center">
        <div className="flex justify-between items-center w-full max-w-[700px] px-4 mb-2">
          <p className="text-lg font-medium">
            {gameState?.game?.current_turn !== null ? (
              <>
                Current Turn:{" "}
                {gameState?.game.players.find(
                  (p) => p.turn_order === gameState.game?.current_turn
                )?.username ?? `Player ${gameState?.game?.current_turn}`}
              </>
            ) : (
              ""
            )}
          </p>
          <div className="mt-2">
            <h3 className="font-semibold mb-1">Scores:</h3>
            <ul className="text-sm space-y-1">
              {gameState?.game.players.map((player) => (
                <li key={player.user_id}>
                  {player.username}: {player.score}
                </li>
              ))}
            </ul>
          </div>
          <Button variant="destructive" onClick={handleQuitGame}>
            Quit Game
          </Button>
        </div>
        <div className="w-full max-w-[500px] md:max-w-[700px] lg:max-w-[900px]">
          {gameState && (
            <Grid
              gameID={gameState.game?.game_id}
              boxes={gameState.grids}
              userColors={userColors}
              boardSize={gameState.game?.board_size}
              userID={user.userID}
              handleClick={handleClick}
            />
          )}
        </div>
      </div>

      {/* Right side: PlayerLobby + Chatbox in column */}
      <div className="w-full md:w-1/3 flex flex-col gap-4">
        <div className="flex-1 overflow-y-auto">
          <PlayerLobby />
        </div>
        <div className="h-64">
          {gameState?.game?.session_id && (
            <Chatbox sessionID={gameState.game.session_id} />
          )}
        </div>
      </div>
      <Dialog open={!!gameState?.game?.winner}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Game Over</DialogTitle>
          </DialogHeader>
          <div className="text-center py-4">
            <p className="text-lg font-semibold">
              {winnerPlayer
                ? `Congratulations to ${winnerPlayer.username}!`
                : "We have a winner!"}
            </p>
          </div>
          <div className="mt-4">
            <h4 className="font-semibold mb-2">Final Scores:</h4>
            <ul className="text-sm space-y-1">
              {gameState?.game?.players.map((player) => (
                <li key={player.user_id}>
                  {player.username}: {player.score}
                </li>
              ))}
            </ul>
          </div>
          <DialogFooter>
            <Button
              onClick={() => {
                setWinnerId(null);
                setWinnerUsername(null);
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
