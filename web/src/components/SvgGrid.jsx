import { useEffect, useState } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import { useNavigate } from "react-router-dom";
import { Toaster, toast } from "sonner";

// eslint-disable-next-line react/prop-types
const Grid = ({ gameID }) => {
  const [boxes, setBoxes] = useState([]);
  const boxSize = 50;
  const [currentTurnPlayerId, setCurrentTurnPlayerId] = useState(null);
  const { socket, subscribe } = useWebSocket();
  const { user } = useUser();
  const navigate = useNavigate();
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    if (!socket) return;

    const unsubscribe = subscribe((message) => {
      if (message.type === "game:state") {
        console.log("Received game:state message", message);
        setBoxes(message.payload.grids);
        setCurrentTurnPlayerId(message.payload.game.CurrentTurn);
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

      <svg width={boxSize * 6} height={boxSize * 6}>
        {boxes.map((box) => {
          const {
            BoxId,
            Row,
            Col,
            top_edge,
            left_edge,
            right_edge,
            bottom_edge,
            completed,
            completed_by,
          } = box;
          const x = Col * boxSize;
          const y = Row * boxSize;
          const completedText = completed_by;
          return (
            <g key={BoxId}>
              {completed && (
                <>
                  <rect
                    x={x}
                    y={y}
                    width={boxSize}
                    height={boxSize}
                    fill="transparent"
                  />

                  <text
                    x={x + boxSize / 2}
                    y={y + boxSize / 2}
                    textAnchor="middle"
                    alignmentBaseline="middle"
                    fontSize="12"
                    fill="black"
                  >
                    {completedText}
                  </text>
                </>
              )}

              {/* Top Edge */}
              <line
                x1={x}
                y1={y}
                x2={x + boxSize}
                y2={y}
                className={top_edge ? "active" : "inactive"}
                strokeWidth="2"
                onClick={() =>
                  handleClick(
                    parseInt(gameID),
                    parseInt(user.userID),
                    Row,
                    Col,
                    "top_edge"
                  )
                }
                style={{ cursor: "pointer" }}
              />
              {/* Left Edge */}
              <line
                x1={x}
                y1={y}
                x2={x}
                y2={y + boxSize}
                className={left_edge ? "active" : "inactive"}
                strokeWidth="2"
                onClick={() =>
                  handleClick(
                    parseInt(gameID),
                    parseInt(user.userID),
                    Row,
                    Col,
                    "left_edge"
                  )
                }
                style={{ cursor: "pointer" }}
              />
              {/* Right Edge */}
              <line
                x1={x + boxSize}
                y1={y}
                x2={x + boxSize}
                y2={y + boxSize}
                className={right_edge ? "active" : "inactive"}
                strokeWidth="2"
                onClick={() =>
                  handleClick(
                    parseInt(gameID),
                    parseInt(user.userID),
                    Row,
                    Col,
                    "right_edge"
                  )
                }
                style={{ cursor: "pointer" }}
              />
              {/* Bottom Edge */}
              <line
                x1={x}
                y1={y + boxSize}
                x2={x + boxSize}
                y2={y + boxSize}
                className={bottom_edge ? "active" : "inactive"}
                strokeWidth="2"
                onClick={() =>
                  handleClick(
                    parseInt(gameID),
                    parseInt(user.userID),
                    Row,
                    Col,
                    "bottom_edge"
                  )
                }
                style={{ cursor: "pointer" }}
              />
            </g>
          );
        })}
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
            <h2>You win!</h2>
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
