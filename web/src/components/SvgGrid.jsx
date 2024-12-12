import { useEffect, useState } from "react";
import { useUser } from "../UserContext"; // Import the useUser hook
import { useWebSocket } from "../WebSocketContext"; // Import the WebSocketContext hook
import { useNavigate } from "react-router-dom"; // Import useNavigate for navigation
import { Toaster, toast } from "sonner";

// eslint-disable-next-line react/prop-types
const Grid = ({ gameID }) => {
  const [boxes, setBoxes] = useState([]);
  const boxSize = 50; // Size of each box in the grid
  const ws = useWebSocket(); // Use WebSocket context
  const { user } = useUser();
  const navigate = useNavigate(); // For navigation
  const [showModal, setShowModal] = useState(false); // State for modal visibility

  useEffect(() => {
    if (!ws) return;

    ws.onopen = () => {
      console.log("WebSocket connected");
      fetchGridData();
    };

    const fetchGridData = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          JSON.stringify({
            type: "get_grids",
            payload: {
              gameID: parseInt(gameID),
            },
          })
        );
      } else {
        console.log("WebSocket not ready, retrying...");
        setTimeout(fetchGridData, 500);
      }
    };

    fetchGridData();

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "new_grids") {
        setBoxes(message.payload);
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

      if (message.type === "quit_game") {
        console.log("User Quit Game");

        setShowModal(true);
      }
    };

    ws.onclose = (event) => {
      console.log("WebSocket disconnected", event);
    };

    ws.onerror = (error) => {
      console.error("WebSocket error", error);
    };

    return () => {
      ws.removeEventListener("message", ws.onmessage);
    };
  }, [gameID, ws]);

  const handleClick = (gameId, playerId, row, col, edge) => {
    if (ws) {
      const payload = {
        gameId,
        playerId,
        row,
        col,
        edge,
      };

      console.log(payload);

      ws.send(
        JSON.stringify({
          type: "make_move",
          payload,
        })
      );
    }
  };

  const handleGoHome = () => {
    navigate("/Home"); // Navigate back to the home page using useNavigate
  };

  const handleQuitGame = () => {
    if (ws) {
      const payload = {
        gameId: parseInt(gameID),
        playerId: parseInt(user.userID),
      };

      ws.send(
        JSON.stringify({
          type: "quit_game",
          payload,
        })
      );
    }

    navigate("/home");
  };

  return (
    <div>
      <Toaster position="top-right" richColors />
      <svg
        width={boxSize * 6}
        height={boxSize * 6}
        style={{ border: "1px solid black" }}
      >
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
                    fill="transparent" // Keep the box transparent
                  />
                  {/* Text showing who completed the box */}
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
                } // Replace 1 with the actual playerId
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
                } // Replace 1 with the actual playerId
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
                } // Replace 1 with the actual playerId
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
                } // Replace 1 with the actual playerId
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
