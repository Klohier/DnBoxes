/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { StrictMode, Suspense } from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import "./styles/globals.css";

// Import the generated route tree
import { routeTree } from "./routeTree.gen";
// import { useAuth } from "./hooks/useAuth";
import { useAuth } from "./AuthContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { AuthProvider } from "./AuthContext";
import { WebSocketProvider } from "./WebSocketContext";

const queryClient = new QueryClient();

// Create a new router instance
const router = createRouter({
  routeTree,
  context: {
    authentication: undefined!,
    queryClient: undefined!,
  },
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

function AppRouter() {
  const authentication = useAuth();

  if (authentication.loading && !authentication.user) {
    return <div>Loading...</div>;
  }

  return (
    <RouterProvider router={router} context={{ authentication, queryClient }} />
  );
}

// Render the app
const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);

  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <WebSocketProvider>
            <Suspense fallback={<div>Loading...</div>}>
              <AppRouter />
            </Suspense>
          </WebSocketProvider>
        </AuthProvider>

        <ReactQueryDevtools />
      </QueryClientProvider>
    </StrictMode>
  );
}
