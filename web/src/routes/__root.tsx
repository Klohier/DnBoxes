import { AuthContextType } from "@/AuthContext";
import { QueryClient } from "@tanstack/react-query";
import { WebSocketProvider } from "./WebSocketContext";

import {
  createRootRoute,
  createRootRouteWithContext,
  Link,
  Outlet,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
// import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
// import { AuthContext } from "@/hooks/useAuth";
// const queryClient = new QueryClient();

type RouterContext = {
  authentication: AuthContextType;
  queryClient: QueryClient;
};
export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => (
    <>
      {/* <QueryClientProvider client={queryClient}> */}
      <div className="p-2 flex gap-2">
        <Link to="/" className="[&.active]:font-bold">
          Home
        </Link>{" "}
        <Link to="/about" className="[&.active]:font-bold">
          About
        </Link>
      </div>
      <hr />
      <Outlet />
      <TanStackRouterDevtools />
      {/* </QueryClientProvider> */}
    </>
  ),
});
