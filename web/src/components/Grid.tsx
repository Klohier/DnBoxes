import { useEffect, useState } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import { useNavigate } from "react-router-dom";
import { Toaster, toast } from "sonner";
import Box from "./Box";

const Grid = ({
  gameID,
  boxes,
  userColors,
  boardSize,
  userID,
  handleClick,
}) => {
  const boxSize = 50;
  const [showModal, setShowModal] = useState(false);
  const navigate = useNavigate();

  const handleGoHome = () => {
    setShowModal(false);
    console.log("Navigating to /home");
    navigate("/home");
  };

  return (
    <div>
      {/* Show current turn */}
      {/* {currentTurnPlayerId !== null && (
        <div style={{ marginBottom: "10px", fontSize: "18px" }}>
          Current Turn: Player {currentTurnPlayerId}
        </div>
      )} */}
      <div className="w-full aspect-square border rounded-lg">
        <svg
          className="w-full h-full"
          viewBox={`-5 -5 ${boxSize * boardSize + 10} ${
            boxSize * boardSize + 10
          }`}
          preserveAspectRatio="xMidYMid meet"
        >
          {boxes.map((box) => (
            <Box
              key={box.box_id}
              box={box}
              userColors={userColors}
              onEdgeClick={handleClick}
              currentUserId={userID}
              gameID={parseInt(gameID)}
              boxSize={boxSize}
              boardSize={boardSize}
            />
          ))}
        </svg>
      </div>
      {/* <button onClick={handleQuitGame} style={{ marginTop: "10px" }}>
        Quit Game
      </button> */}
      {showModal && (
        <div
          style={{
            position: "fixed",
            top: "0",
            left: "0",
            width: "100%",
            height: "100%",
            backgroundColor: "rgba(0,0,0,0.5)",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            zIndex: "1000",
          }}
        >
          <div
            style={{
              backgroundColor: "white",
              padding: "20px",
              borderRadius: "10px",
              textAlign: "center",
              width: "300px",
            }}
          >
            <h2>
              {parseInt(user.userID) === winnerId ? "You win!" : "You lost!"}
            </h2>
            <h3 style={{ marginTop: "10px" }}>Final Scores</h3>
            <ul style={{ listStyle: "none", padding: 0 }}>
              {playerScores.map((player) => (
                <li
                  key={player.user_id}
                  style={{
                    fontWeight:
                      parseInt(player.user_id) === winnerId ? "bold" : "normal",
                    color: userColors[player.user_id],
                    marginBottom: "5px",
                  }}
                >
                  Player {player.user_id}
                  {player.username ? ` (${player.username})` : ""}:{" "}
                  {player.score}
                </li>
              ))}
            </ul>
            <button onClick={handleGoHome} style={{ marginTop: "20px" }}>
              Go to Home
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Grid;
