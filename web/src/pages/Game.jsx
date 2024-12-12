import { useParams } from "react-router-dom";
import Grid from "../components/SvgGrid";
import Chatbox from "../components/ChatBox";

const Game = () => {
  const { gameID } = useParams(); // Extract gameID from the route

  return (
    <div>
      <h1>Game Page</h1>
      <p>Game ID: {gameID}</p>
      <Grid gameID={gameID} />
      <Chatbox gameID={gameID}></Chatbox>
      {/* Add game-related logic and UI here */}
    </div>
  );
};

export default Game;
