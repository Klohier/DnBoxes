import { useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import { Link } from "@tanstack/react-router";

export function Header() {
  const auth = useAuth();

  return (
    <header className="bg-gray-800 border-b border-gray-700 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between min-w-0">
        <div className="flex items-center gap-6 min-w-0">
          <Link
            to="/"
            className="text-xl font-bold text-white hover:text-gray-300 transition-colors whitespace-nowrap"
          >
            Dots & Boxes
          </Link>

          <nav className="hidden sm:flex gap-4">
            <Link
              to="/play"
              className="text-gray-300 hover:text-white transition-colors [&.active]:text-white [&.active]:font-semibold"
            >
              Play
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

        <div className="flex items-center gap-3 shrink-0">
          {auth.loading ? (
            <Button disabled variant="outline" size="sm">
              Loading...
            </Button>
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
            <Link to="/login">
              <Button variant="default" size="sm">
                Login
              </Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
