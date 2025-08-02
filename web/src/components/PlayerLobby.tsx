import { useState, useEffect } from "react";
import { useWebSocket } from "../WebSocketContext";
import { useUser } from "../UserContext";
import { useNavigate } from "@tanstack/react-router";
import { Message, Player, InvitePayload } from "@/types/websocket";
import PlayerList from "./PlayerList";
import SendInviteModal from "./SendInviteModal";
import IncomingInviteModal from "./IncomingInviteModal";
import { useAuth } from "@/AuthContext";

const PlayerLobby = () => {
  const [players, setPlayers] = useState<Player[]>([]);

  const [boardSize, setBoardSize] = useState(5);
  const [incomingInvite, setIncomingInvite] = useState<InvitePayload | null>(
    null
  );
  const { user } = useAuth();
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>();
  const { send, subscribe, connected } = useWebSocket();
  const navigate = useNavigate();

  useEffect(() => {
    if (connected) {
      send({ type: "player:get", payload: "lobby" });
    }
  }, [connected]);

  useEffect(() => {
    const unsubscribe = subscribe((message: Message) => {
      if (message.type === "player:get" && message.payload) {
        setPlayers(message.payload as Player[]);
      }
      if (message.type === "invite:new") setIncomingInvite(message.payload);
      if (message.type === "game:new")
        void navigate({
          to: "/game/$gameID",
          params: { gameID: message.payload.gameID },
        });
    });

    return () => {
      unsubscribe();
    };
  }, [subscribe, navigate]);

  const handlePlayerClick = (player: Player) => {
    setSelectedPlayer(player);
  };

  const handleSendGameInvite = () => {
    if (selectedPlayer && user) {
      send({
        type: "invite:new",
        payload: {
          senderID: user.userID,
          senderName: user.username,
          receiverID: selectedPlayer.user_id,
          receiverName: selectedPlayer.username,
          timestamp: new Date().toISOString(),
          board_size: boardSize,
        },
      });
      setSelectedPlayer(null);
    }
  };

  const handleAcceptInvite = () => {
    if (incomingInvite && user) {
      send({
        type: "invite:accept",
        payload: {
          playerID: user.userID,
          senderID: incomingInvite.senderID,
          board_size: incomingInvite.board_size,
        },
      });
      setIncomingInvite(null);
    }
  };

  const handleDeclineInvite = () => {
    if (incomingInvite && user) {
      send({
        type: "invite:decline",
        payload: {
          inviterID: incomingInvite.senderID,
        },
      });
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
        onClose={() => {
          setSelectedPlayer(null);
        }}
      />

      <IncomingInviteModal
        incomingInvite={incomingInvite}
        onAccept={handleAcceptInvite}
        onDecline={handleDeclineInvite}
        onClose={() => {
          setIncomingInvite(null);
        }}
      />
    </div>
  );
};

export default PlayerLobby;
