import { useState, useEffect } from "react";
import { useWebSocket } from "../WebSocketContext";
import { useUser } from "../UserContext";
import { useNavigate } from "react-router-dom";

const PlayerList = () => {
  const [players, setPlayers] = useState([]);
  const [selectedPlayer, setSelectedPlayer] = useState(null);
  const [boardSize, setBoardSize] = useState(5);
  const [incomingInvite, setIncomingInvite] = useState(null);
  const { user } = useUser();
  const ws = useWebSocket();
  const navigate = useNavigate();

  useEffect(() => {
    if (!ws) return;

    ws.onopen = () => {
      console.log("WebSocket connected");
      ws.send(JSON.stringify({ type: "player:get" }));
    };

    console.log(user);

    if (user?.gameID) {
      navigate(`/game/${user.gameID}`);
      return;
    }

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === "player:get") {
        setPlayers(message.payload);
      }

      if (message.type === "invite:new") {
        setIncomingInvite(message.payload);
      }

      if (message.type === "game:new") {
        const { gameID } = message.payload;
        console.log(`Redirecting to game page with ID: ${gameID}`);
        navigate(`/game/${gameID}`);
      }
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    return () => {
      ws.removeEventListener("message", ws.onmessage);
    };
  }, [navigate, user, ws]);

  const handlePlayerClick = (player) => {
    setSelectedPlayer(player);
  };

  const handleCloseModal = () => {
    setSelectedPlayer(null);
  };

  const handleBoardSizeChange = (e) => {
    setBoardSize(parseInt(e.target.value));
  };

  const handleSendGameInvite = () => {
    if (selectedPlayer) {
      // Send a game invite to the selected player
      ws.send(
        JSON.stringify({
          type: "invite:new",
          payload: {
            senderID: user.userID,
            senderName: user.username, // The ID of the user sending the invite
            receiverID: selectedPlayer.userID,
            receiverName: selectedPlayer.username, // The ID of the player receiving the invite
            timestamp: new Date().toISOString(), // Optional: to track when the invite was sent
            board_size: boardSize,
          },
        })
      );
      console.log(`Game invite sent to ${selectedPlayer.username}`);
      handleCloseModal(); // Close modal after sending invite
    }
  };

  const handleAcceptInvite = () => {
    if (incomingInvite) {
      ws.send(
        JSON.stringify({
          type: "invite:accept",
          payload: {
            playerID: user.userID,
            senderID: incomingInvite.senderID,
            board_size: incomingInvite.board_size,
          },
        })
      );
      console.log(`Accepted invite from ${incomingInvite.username}`);
      setIncomingInvite(null); // Close the modal
    }
  };

  const handleDeclineInvite = () => {
    if (incomingInvite) {
      ws.send(
        JSON.stringify({
          type: "invite:decline",
          payload: {
            inviterID: incomingInvite.inviterID,
          },
        })
      );
      console.log(`Declined invite from ${incomingInvite.inviterName}`);
      setIncomingInvite(null); // Close the modal
    }
  };

  return (
    <div>
      <h3>Players</h3>
      <ul>
        {players.map((player, index) => (
          <li
            key={index}
            onClick={() => handlePlayerClick(player)}
            style={{
              cursor: "pointer",
              padding: "5px",
              borderBottom: "1px solid #ccc",
            }}
          >
            {player.username} (ID: {player.userID})
          </li>
        ))}
      </ul>

      {/* Modal for sending game invite */}
      {selectedPlayer && (
        <div
          style={{
            position: "fixed",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            backgroundColor: "white",
            padding: "20px",
            border: "1px solid #ccc",
            boxShadow: "0 4px 8px rgba(0, 0, 0, 0.2)",
            zIndex: 1000,
            width: "300px",
          }}
        >
          <h4>Send Game Invite</h4>
          <p>
            Are you sure you want to send a game invite to{" "}
            {selectedPlayer.username}?
          </p>
          <div>
            <label htmlFor="boardSize">Board Size: </label>
            <input
              type="number"
              id="boardSize"
              value={boardSize}
              onChange={handleBoardSizeChange}
              min="1"
              max="100"
              style={{ marginLeft: "10px", width: "50px" }}
            />
          </div>
          <button onClick={handleSendGameInvite}>Send Invite</button>
          <button onClick={handleCloseModal}>Cancel</button>
        </div>
      )}

      {/* Overlay background for the modal */}
      {selectedPlayer && (
        <div
          onClick={handleCloseModal}
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%",
            backgroundColor: "rgba(0, 0, 0, 0.5)",
            zIndex: 999,
          }}
        />
      )}

      {/* Modal for incoming invites */}
      {incomingInvite && (
        <div
          style={{
            position: "fixed",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            backgroundColor: "white",
            padding: "20px",
            border: "1px solid #ccc",
            boxShadow: "0 4px 8px rgba(0, 0, 0, 0.2)",
            zIndex: 1000,
            width: "300px",
          }}
        >
          <h4>Game Invite</h4>
          <p>
            {incomingInvite.inviterName} has invited you to a game with a board
            size of {incomingInvite.board_size}.
          </p>
          <button onClick={handleAcceptInvite}>Accept</button>
          <button onClick={handleDeclineInvite}>Decline</button>
        </div>
      )}

      {/* Overlay background for the modal */}
      {incomingInvite && (
        <div
          onClick={() => setIncomingInvite(null)}
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%",
            backgroundColor: "rgba(0, 0, 0, 0.5)",
            zIndex: 999,
          }}
        />
      )}
    </div>
  );
};

export default PlayerList;
