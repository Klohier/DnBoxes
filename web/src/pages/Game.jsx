import { useParams } from "react-router-dom";
import axios from "axios";
import { useEffect, useState } from "react";
// import { useNavigate } from "react-router-dom";
import Grid from "../components/Grid";
import Chatbox from "../components/ChatBox";
// import { useWebSocket } from "../WebSocketContext";
import PlayerLobby from "../components/PlayerLobby";

const Game = () => {
  const { gameID } = useParams();
  const [sessionID, setSessionID] = useState(null);
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";

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

  return (
    <div>
      <p>Game ID: {gameID}</p>
      <Grid gameID={gameID} />
      <Chatbox sessionID={sessionID}></Chatbox>
      <PlayerLobby></PlayerLobby>
    </div>
  );
};

export default Game;
