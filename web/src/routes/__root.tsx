import { AuthContextType, useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
// import { GameStatePayload } from "@/types/websocket";
import { QueryClient } from "@tanstack/react-query";

import {
  createRootRouteWithContext,
  Link,
  Outlet,
  // useRouteContext,
  // useRouter,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
// import axios from "axios";
import { useEffect } from "react";

interface RouterContext {
  authentication: AuthContextType;
  queryClient: QueryClient;
}
export const Route = createRootRouteWithContext<RouterContext>()({
  component: Root,
});

function Root() {
  // const router = useRouter();
  const auth = useAuth();
  // const [loading, setLoading] = useState<boolean>();

  // async function handleCreateBotGame() {
  //   if (!auth.isAuthenticated) {
  //     alert("User not authenticated");
  //     return;
  //   }
  //   setLoading(true);

  //   try {
  //     const response = await axios.post<GameStatePayload>(
  //       `/api/v1/games/create-bot-game`,
  //       {
  //         human_player_id: auth.user?.userID,
  //         board_size: 5,
  //         num_bots: 1,
  //         session_id: Date.now() - Math.floor(Math.random() * 100000),
  //       },
  //       {
  //         headers: { "Content-Type": "application/json" },
  //       }
  //     );

  //     const game = response.data;
  //     await router.navigate({
  //       to: "/game/$gameID",
  //       params: { gameID: String(game.game.game_id) },
  //     });
  //   } catch (error) {
  //     alert("Error creating bot game: " + String(error));
  //   } finally {
  //     setLoading(false);
  //   }
  // }

  useEffect(() => {
    console.log("Auth Debug:", {
      loading: auth.loading,
      isAuthenticated: auth.isAuthenticated,
      user: auth.user,
    });
  }, [auth.loading]);

  return (
    <>
      <div className="p-2 flex gap-2">
        <Link to="/" className="[&.active]:font-bold">
          Home
        </Link>
        <Link to="/about" className="[&.active]:font-bold">
          About
        </Link>

        {auth.loading ? (
          <div className="flex gap-2">
            <Button disabled variant="outline">
              Loading...
            </Button>
            <Button disabled variant="outline">
              Loading...
            </Button>
          </div>
        ) : auth.isAuthenticated ? (
          <>
            <Button onClick={auth.logout} variant={"destructive"}>
              Logout
            </Button>
            {/* <Button
              onClick={() => void handleCreateBotGame()}
              disabled={loading}
            >
              {loading ? "Creating Game..." : "Create Bot Game"}
            </Button> */}
          </>
        ) : null}
      </div>
      <hr />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
}
