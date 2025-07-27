import {
  createFileRoute,
  redirect,
  useRouteContext,
} from "@tanstack/react-router";

import Chatbox from "../../components/ChatBox";
import PlayerLobby from "../../components/PlayerLobby";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  // const { authentication } = useRouteContext({ from: "/_authenticated/" });
  return (
    <>
      <PlayerLobby />
      <Chatbox sessionID={1} />
    </>
  );
}
