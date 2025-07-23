import { useState, useEffect } from "react";
import { useWebSocket } from "../WebSocketContext";
import { useUser } from "../UserContext";
import { useNavigate } from "react-router-dom";
import { Message, Player } from "@/types/websocket";
import PlayerList from "./PlayerList";
import SendInviteModal from "./SendInviteModal";
import IncomingInviteModal from "./IncomingInviteModal";

const PlayerLobby = () => {
  const [players, setPlayers] = useState([]);

  const [boardSize, setBoardSize] = useState(5);
  const [incomingInvite, setIncomingInvite] = useState();
  const { user } = useUser();
  const [selectedPlayer, setSelectedPlayer] = useState<Player>();
  const { socket, subscribe } = useWebSocket();
  const navigate = useNavigate();

  // useEffect(() => {
  //   if (user?.gameID) {
  //     navigate(`/game/${user.gameID}`);
  //   }
  // }, [user?.gameID, navigate]);

  useEffect(() => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("No WebSocket connection available");
      return;
    }

    const unsubscribe = subscribe((message) => {
      if (message.type === "player:get") setPlayers(message.payload);
      if (message.type === "invite:new") setIncomingInvite(message.payload);
      if (message.type === "game:new")
        navigate(`/game/${message.payload.gameID}`);
    });

    if (socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: "player:get" }));
    } else {
      socket.onopen = () => {
        socket.send(JSON.stringify({ type: "player:get" }));
      };
    }

    return () => { unsubscribe(); };
  }, [socket, subscribe, navigate]);

  const handlePlayerClick = (player) => {
    setSelectedPlayer(player);
  };

  const handleSendGameInvite = () => {
    if (selectedPlayer && socket && user) {
      socket.send(
        JSON.stringify({
          type: "invite:new",
          payload: {
            senderID: user.userID,
            senderName: user.username,
            receiverID: selectedPlayer.user_id,
            receiverName: selectedPlayer.username,
            timestamp: new Date().toISOString(),
            board_size: boardSize,
          },
        })
      );
      setSelectedPlayer(null);
    }
  };

  const handleAcceptInvite = () => {
    if (incomingInvite && socket && user) {
      socket.send(
        JSON.stringify({
          type: "invite:accept",
          payload: {
            playerID: user.userID,
            senderID: incomingInvite.senderID,
            board_size: incomingInvite.board_size,
          },
        })
      );
      setIncomingInvite(null);
    }
  };

  const handleDeclineInvite = () => {
    if (incomingInvite && socket && user) {
      socket.send(
        JSON.stringify({
          type: "invite:decline",
          payload: {
            inviterID: incomingInvite.inviterID,
          },
        })
      );
      setIncomingInvite(null);
    }
  };

  return (
    <div>
      <h3>Players</h3>
      <PlayerList players={players} onPlayerClick={handlePlayerClick} />

      <SendInviteModal
        selectedPlayer={selectedPlayer}
        boardSize={boardSize}
        onBoardSizeChange={setBoardSize}
        onSendInvite={handleSendGameInvite}
        onClose={() => { setSelectedPlayer(null); }}
      />

      <IncomingInviteModal
        incomingInvite={incomingInvite}
        onAccept={handleAcceptInvite}
        onDecline={handleDeclineInvite}
        onClose={() => { setIncomingInvite(null); }}
      />
    </div>
  );
};

export default PlayerLobby;
