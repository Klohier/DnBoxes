import {
  createFileRoute,
  redirect,
  useRouteContext,
} from "@tanstack/react-router";

import Chatbox from "../components/ChatBox";
import PlayerLobby from "../components/PlayerLobby";

export const Route = createFileRoute("/")({
  beforeLoad: ({ context }) => {
    const { isAuthenticated } = context.authentication;
    // await context.queryClient.invalidateQueries({ queryKey: ["me"] });

    if (!isAuthenticated) {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect({
        to: "/login",
      });
    }
  },
  component: Index,
});

function Index() {
  const { authentication } = useRouteContext({ from: "/" });
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
