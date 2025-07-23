import { useParams } from "react-router-dom";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Grid from "../components/Grid";
import Chatbox from "../components/ChatBox";
import { useWebSocket } from "../WebSocketContext";
import PlayerLobby from "../components/PlayerLobby";
import { useUser } from "@/UserContext";
import { Toaster, toast } from "sonner";
import { Box, GamePlayer, Message } from "@/types/websocket";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

const Game = () => {
  const gameID = Number(useParams<{ gameID: string }>().gameID);
  const [boxes, setBoxes] = useState<Box[]>([]);
  const [sessionID, setSessionID] = useState<number>();
  const { user } = useUser();
  const { socket, subscribe } = useWebSocket();
  const navigate = useNavigate();
  const [winnerUsername, setWinnerUsername] = useState<string | null>(null);
  const [players, setPlayers] = useState<GamePlayer[]>([]);
  const [userColors, setUserColors] = useState<Record<string, string>>({});
  const [boardSize, setBoardSize] = useState<number>(5);
  const [winnerId, setWinnerId] = useState<number | null>(null);
  const [playerScores, setPlayerScores] = useState([]);
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";
  const [currentTurnPlayerId, setCurrentTurnPlayerId] = useState<
    number | undefined
  >();

  useEffect(() => {
    if (!socket) return;

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
        const game = message.payload.game;
        const players = game.players;
        setBoxes(message.payload.grids);
        setCurrentTurnPlayerId(game.current_turn);
        setSessionID(game.session_id);
        setBoardSize(game.board_size);

        if (players) {
          const colorMap: Record<GamePlayer["user_id"], string> = {};
          players.forEach((player, index) => {
            colorMap[player.user_id] = colors[index % colors.length];
          });
          setUserColors(colorMap);
          setPlayers(players);
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
        // const winnerId = message.payload.winnerId;
        // const winnerUsername = message.payload.winnerUsername;

        // setWinnerId(winnerId);
        // setWinnerUsername(winnerUsername);
      }
    });

    if (socket.readyState === WebSocket.OPEN) {
      socket.send(
        JSON.stringify({
          type: "game:state",
          payload: { gameID: gameID },
        })
      );
    }

    return () => unsubscribe();
  }, [socket, subscribe, gameID]);

  const handleQuitGame = () => {
    if (sessionID && user?.userID) {
      socket?.send(
        JSON.stringify({
          type: "game:quit",
          payload: {
            gameId: gameID,
            playerId: user.userID,
            session_id: sessionID,
          },
        })
      );
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
    if (socket) {
      const payload = {
        gameId,
        playerId,
        row,
        col,
        edge,
      };

      console.log(payload);

      socket.send(
        JSON.stringify({
          type: "game:move",
          payload,
        })
      );
    }
  };

  return (
    <div className="flex flex-col md:flex-row h-screen p-4 gap-4">
      <Toaster position="top-right" richColors />

      {/* Grid section */}
      <div className="w-full md:w-2/3 flex flex-col items-center">
        <div className="flex justify-between items-center w-full max-w-[700px] px-4 mb-2">
          <p className="text-lg font-medium">
            {currentTurnPlayerId !== null ? (
              <>
                Current Turn:{" "}
                {players.find((p) => p.turn_order === currentTurnPlayerId)
                  ?.username ?? `Player ${currentTurnPlayerId}`}
              </>
            ) : (
              ""
            )}
          </p>
          <div className="mt-2">
            <h3 className="font-semibold mb-1">Scores:</h3>
            <ul className="text-sm space-y-1">
              {players.map((player) => (
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
          {user && (
            <Grid
              gameID={gameID}
              boxes={boxes}
              userColors={userColors}
              boardSize={boardSize}
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
          <Chatbox
            sessionID={sessionID}
            // setCurrentTurnPlayerId={setCurrentTurnPlayerId}
          />
        </div>
      </div>
      <Dialog open={winnerId !== null}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Game Over</DialogTitle>
          </DialogHeader>
          <div className="text-center py-4">
            <p className="text-lg font-semibold">
              {winnerUsername
                ? `Congratulations to ${winnerUsername}!`
                : "We have a winner!"}
            </p>
          </div>
          <div className="mt-4">
            <h4 className="font-semibold mb-2">Final Scores:</h4>
            <ul className="text-sm space-y-1">
              {players.map((player) => (
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
                navigate("/home");
              }}
            >
              Return Home
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default Game;
