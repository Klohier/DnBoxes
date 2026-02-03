import { AuthContextType, useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import { QueryClient } from "@tanstack/react-query";

import {
  createRootRouteWithContext,
  Link,
  Outlet,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

interface RouterContext {
  authentication: AuthContextType;
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: Root,
});

function Root() {
  const auth = useAuth();

  return (
    <>
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link
              to="/"
              className="text-xl font-bold text-white hover:text-gray-300 transition-colors"
            >
              Dots & Boxes
            </Link>

            {/* Navigation Links */}
            <nav className="flex gap-4">
              <Link
                to="/"
                className="text-gray-300 hover:text-white transition-colors [&.active]:text-white [&.active]:font-semibold"
              >
                Home
              </Link>
              <Link
                to="/leaderboard"
                className="text-gray-300 hover:text-white transition-colors [&.active]:text-white [&.active]:font-semibold"
              >
                Leaderboard
              </Link>
              {auth.isAuthenticated && (
                <Link
                  to="/history"
                  className="text-gray-300 hover:text-white transition-colors [&.active]:text-white [&.active]:font-semibold"
                >
                  History
                </Link>
              )}
              <Link
                to="/about"
                className="text-gray-300 hover:text-white transition-colors [&.active]:text-white [&.active]:font-semibold"
              >
                About
              </Link>
            </nav>
          </div>

          {/* Auth Section */}
          <div className="flex items-center gap-3">
            {auth.loading ? (
              <div className="flex gap-2">
                <Button disabled variant="outline" size="sm">
                  Loading...
                </Button>
              </div>
            ) : auth.isAuthenticated ? (
              <>
                <span className="text-gray-400 text-sm hidden sm:inline">
                  Welcome back!
                </span>
                <Button onClick={auth.logout} variant="destructive" size="sm">
                  Logout
                </Button>
              </>
            ) : (
              <Button variant="default" size="sm">
                Login
              </Button>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <TanStackRouterDevtools />
    </>
  );
}
