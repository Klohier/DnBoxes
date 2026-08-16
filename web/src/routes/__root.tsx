import { AuthContextType } from "@/AuthContext";

import { QueryClient } from "@tanstack/react-query";
import { Header } from "@/components/Header";

import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { HeadContent } from "@tanstack/react-router";

interface RouterContext {
  authentication: AuthContextType;
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: Root,
});

function Root() {
  return (
    <>
      <HeadContent />
      <Header />

      <main className="flex-1">
        <Outlet />
      </main>

      <TanStackRouterDevtools />
    </>
  );
}
