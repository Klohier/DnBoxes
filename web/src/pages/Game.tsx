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

const Game = () => {
  const { gameID } = useParams();
  const [sessionID, setSessionID] = useState(null);
  const { user } = useUser();
  const { socket, subscribe } = useWebSocket();
  const navigate = useNavigate();

  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";
  const [currentTurnPlayerId, setCurrentTurnPlayerId] = useState<number | null>(
    null
  );

  useEffect(() => {
    const fetchGameState = async () => {
      try {
        const response = await axios.get(
          `http://${apiUrl}/api/v1/games/${gameID}/state`
        );
        if (response.data) {
          setSessionID(response.data.game.session_id);
        } else {
          console.error("No game state found for game", gameID);
        }
      } catch (err) {
        console.error("Failed to fetch game state for game", gameID, err);
      }
    };

    fetchGameState();
  }, [gameID]);

  const handleQuitGame = () => {
    if (sessionID && user?.userID) {
      socket?.send(
        JSON.stringify({
          type: "game:quit",
          payload: {
            gameId: parseInt(gameID!),
            playerId: user.userID,
            session_id: parseInt(sessionID),
          },
        })
      );
    }
    navigate("/home");
  };

  return (
    <div className="flex flex-col md:flex-row h-screen p-4 gap-4">
      {/* Grid section */}
      <div className="w-full md:w-2/3 flex flex-col items-center">
        <div className="flex justify-between items-center w-full max-w-[700px] px-4 mb-2">
          <p className="text-lg font-medium">
            {currentTurnPlayerId !== null
              ? `Current Turn: Player ${currentTurnPlayerId}`
              : ""}
          </p>
          <Button variant="destructive" onClick={handleQuitGame}>
            Quit Game
          </Button>
        </div>
        <div className="w-full max-w-[500px] md:max-w-[700px] lg:max-w-[900px]">
          <Grid gameID={gameID} />
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
            setCurrentTurnPlayerId={setCurrentTurnPlayerId}
          />
        </div>
      </div>
    </div>
  );
};

export default Game;
