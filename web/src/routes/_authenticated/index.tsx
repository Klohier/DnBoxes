import { createFileRoute } from "@tanstack/react-router";

import Chatbox from "../../components/ChatBox";
import PlayerLobby from "../../components/PlayerLobby";
import { useEffect } from "react";
import { useWebSocket } from "@/WebSocketContext";
import LobbyList from "@/components/Lobby";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  const { send } = useWebSocket();
  const topic = "global:lobbies";
  useEffect(() => {
    send({
      type: "page:join",
      payload: { topic: topic },
    });

    return () => {
      send({
        type: "page:leave",
        payload: { topic: topic },
      });
    };
  }, [send]);

  return (
    <>
      {/* <PlayerLobby /> */}
      <Chatbox sessionID={1} />
      <LobbyList />
    </>
  );
}
