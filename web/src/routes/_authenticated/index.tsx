import {
  createFileRoute,
  redirect,
  useRouteContext,
} from "@tanstack/react-router";

import Chatbox from "../../components/ChatBox";
import PlayerLobby from "../../components/PlayerLobby";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

function Index() {
  const { authentication } = useRouteContext({ from: "/_authenticated/" });
  // const user = authentication.user;
  return (
    <>
      <PlayerLobby></PlayerLobby>
      <Chatbox sessionID={1}></Chatbox>
      <button onClick={authentication.logout} style={{ marginTop: "20px" }}>
        Logout
      </button>
    </>
  );
}
