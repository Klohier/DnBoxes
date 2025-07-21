import { useEffect, useState } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import { useNavigate } from "react-router-dom";
import { Toaster, toast } from "sonner";
import Box from "./Box";

// eslint-disable-next-line react/prop-types
const Grid = ({ gameID }) => {
  const [boxes, setBoxes] = useState([]);
  const [sessionID, setSessionID] = useState([]);
  const boxSize = 50;
  const [currentTurnPlayerId, setCurrentTurnPlayerId] = useState(null);
  const { socket, subscribe } = useWebSocket();
  const { user } = useUser();
  const navigate = useNavigate();
  const [showModal, setShowModal] = useState(false);
  const [userColors, setUserColors] = useState({});
  const [boardSize, setBoardSize] = useState();
  const [winnerId, setWinnerId] = useState(null);
  const [playerScores, setPlayerScores] = useState([]);

  useEffect(() => {
    if (!socket) return;

    const colors = ["red", "blue", "green", "purple", "orange", "pink"];

    const unsubscribe = subscribe((message) => {
      if (message.type === "game:state") {
        console.log("Received game:state message", message);
        setBoxes(message.payload.grids);
        setCurrentTurnPlayerId(message.payload.game.current_turn);
        setSessionID(message.payload.game.session_id);
        setBoardSize(message.payload.game.board_size);
        const players = message.payload.game.players;
        if (players) {
          const colorMap = {};
          players.forEach((player, index) => {
            colorMap[player.user_id] = colors[index % colors.length];
          });
          setUserColors(colorMap);
          setPlayerScores(players);
        }
      }
      if (message.type === "winner_set") {
        const winnerId = message.payload.winnerId;
        setWinnerId(winnerId);
        setShowModal(true);
        toast.success(
          `Player ${message.payload.winnerUsername} has won the game!`
        );
      }

      if (message.type === "your_turn") {
        toast("It's your turn!", {
          description: "Make your move now.",
          type: "info",
        });
      }

      if (message.type === "invalid_move") {
        toast.error(
          "Invalid Move, not your turn or selected already selected block"
        );
      }

      if (message.type === "game:quit") {
        console.log("User Quit Game");
        setShowModal(true);
      }
    });

    if (socket.readyState === WebSocket.OPEN) {
      fetchGridData();
    } else {
      const interval = setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          fetchGridData();
          clearInterval(interval);
        }
      }, 500);
    }

    return () => {
      unsubscribe();
    };
  }, [socket, subscribe, gameID]);

  const fetchGridData = () => {
    socket.send(
      JSON.stringify({
        type: "game:state",
        payload: {
          gameID: parseInt(gameID),
        },
      })
    );
  };

  const handleClick = (gameId, playerId, row, col, edge) => {
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

  const handleGoHome = () => {
    setShowModal(false);
    console.log("Navigating to /home");
    navigate("/home"); // Navigate back to the home page using useNavigate
  };

  const handleQuitGame = () => {
    if (socket) {
      const payload = {
        gameId: parseInt(gameID),
        playerId: parseInt(user.userID),
        session_id: parseInt(sessionID),
      };

      socket.send(
        JSON.stringify({
          type: "game:quit",
          payload,
        })
      );
    }
    console.log("Navigating to /home");
    navigate("/home");
  };

  return (
    <div>
      <Toaster position="top-right" richColors />

      {/* Show current turn */}
      {currentTurnPlayerId !== null && (
        <div style={{ marginBottom: "10px", fontSize: "18px" }}>
          Current Turn: Player {currentTurnPlayerId}
        </div>
      )}

      <svg
        width="100%"
        height="100%"
        viewBox={`-5 -5 ${boxSize * boardSize + 10} ${boxSize * boardSize + 10}`}
        preserveAspectRatio="xMidYMid meet"
      >
        {boxes.map((box) => (
          <Box
            key={box.box_id}
            box={box}
            userColors={userColors}
            onEdgeClick={handleClick}
            currentUserId={parseInt(user.userID)}
            gameID={parseInt(gameID)}
            boxSize={boxSize}
            boardSize={boardSize}
          />
        ))}
      </svg>
      <button onClick={handleQuitGame} style={{ marginTop: "10px" }}>
        Quit Game
      </button>
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
