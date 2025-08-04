import { AuthContextType } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import { GameStatePayload } from "@/types/websocket";
import { QueryClient } from "@tanstack/react-query";

import {
  createRootRouteWithContext,
  Link,
  Outlet,
  // useMatchRoute,
  useRouteContext,
  useRouter,
  // useRouterState,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import axios from "axios";
import { useEffect, useState } from "react";

interface RouterContext {
  authentication: AuthContextType;
  queryClient: QueryClient;
}
export const Route = createRootRouteWithContext<RouterContext>()({
  component: Root,
});

function Root() {
  const router = useRouter();
  const auth = useRouteContext({ from: "__root__" });
  const [loading, setLoading] = useState<boolean>();
  // const apiUrl =
  //   (import.meta.env.VITE_API_URL as string) || "http://localhost:8484";

  useEffect(() => {
    console.log("Auth change", auth.authentication.isAuthenticated);
    if (!auth.authentication.isAuthenticated) {
      void router.navigate({ to: "/login" });
    }
  }, [auth.authentication.isAuthenticated]);

  async function handleCreateBotGame() {
    if (!auth.authentication.isAuthenticated) {
      alert("User not authenticated");
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post<GameStatePayload>(
        `/api/v1/games/create-bot-game`,
        {
          human_player_id: auth.authentication.user?.userID,
          board_size: 5,
          session_id: Date.now() - Math.floor(Math.random() * 100000),
        },
        {
          headers: { "Content-Type": "application/json" },
        }
      );

      const game = response.data;
      await router.navigate({
        to: "/game/$gameID",
        params: { gameID: String(game.game.game_id) },
      });
    } catch (error) {
      alert("Error creating bot game: " + String(error));
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="p-2 flex gap-2">
        <Link to="/" className="[&.active]:font-bold">
          Home
        </Link>
        <Link to="/about" className="[&.active]:font-bold">
          About
        </Link>

        {auth.authentication.isAuthenticated && (
          <>
            <Button
              onClick={auth.authentication.logout}
              variant={"destructive"}
            >
              Logout
            </Button>
            <Button
              onClick={() => void handleCreateBotGame()}
              disabled={loading}
            >
              {loading ? "Creating Game..." : "Create Bot Game"}
            </Button>
          </>
        )}
      </div>
      <hr />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
}
