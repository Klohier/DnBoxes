import { AuthContextType } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import { QueryClient } from "@tanstack/react-query";

import {
  createRootRouteWithContext,
  Link,
  Outlet,
  useMatchRoute,
  useRouteContext,
  useRouter,
  useRouterState,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { useEffect } from "react";

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

  useEffect(() => {
    console.log("Auth change", auth.authentication.isAuthenticated);
    if (!auth.authentication.isAuthenticated) {
      void router.navigate({ to: "/login" });
    }
  }, [auth.authentication.isAuthenticated]);

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
          </>
        )}
      </div>
      <hr />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
}
