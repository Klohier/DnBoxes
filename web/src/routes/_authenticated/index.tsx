import { createFileRoute } from "@tanstack/react-router";

import Chatbox from "../../components/ChatBox";
import PlayerLobby from "../../components/PlayerLobby";
import { Button } from "@/components/ui/button";
import { useEffect } from "react";
import { useWebSocket } from "@/WebSocketContext";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  const { send } = useWebSocket();

  useEffect(() => {
    send({
      type: "page:join",
      payload: { topic: "lobby" },
    });

    return () => {
      send({
        type: "page:leave",
        payload: { topic: "lobby" },
      });
    };
  }, [send]);

  return (
    <>
      <PlayerLobby />
      <Chatbox sessionID={1} />
    </>
  );
}
