import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { fetchGameHistory, GameHistoryEntry } from "@/api/fetchGames";
import { useAuth } from "@/AuthContext";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Play } from "lucide-react";

export const Route = createFileRoute("/_authenticated/history")({
  component: HistoryPage,
});

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function GameCard({
  game,
  currentUserID,
}: {
  game: GameHistoryEntry;
  currentUserID: number;
}) {
  const isWinner = game.winner_id === currentUserID;
  const isDraw = game.winner_id === null;
  const winnerPlayer = game.players.find(
    (p) => p.user_id === game.winner_id,
  );

  let resultLabel: string;
  let resultColor: string;
  if (isDraw) {
    resultLabel = "Draw";
    resultColor = "text-yellow-400";
  } else if (isWinner) {
    resultLabel = "Victory";
    resultColor = "text-green-400";
  } else {
    resultLabel = "Defeat";
    resultColor = "text-red-400";
  }

  return (
    <Card className="bg-gray-800 border-gray-700">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base text-white">
            Game #{game.game_id}
          </CardTitle>
          <span className={`text-sm font-semibold ${resultColor}`}>
            {resultLabel}
          </span>
        </div>
        <div className="flex items-center gap-3 text-xs text-gray-400">
          <span>{game.board_size}x{game.board_size} board</span>
          <span>{formatDate(game.created_at)}</span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-1.5">
          {game.players
            .sort((a, b) => b.score - a.score)
            .map((player) => {
              const isCurrentUser = player.user_id === currentUserID;
              const isPlayerWinner = player.user_id === game.winner_id;
              return (
                <div
                  key={player.user_id}
                  className={`flex items-center justify-between px-2 py-1.5 rounded ${
                    isPlayerWinner
                      ? "bg-green-900/20 border border-green-800/40"
                      : "bg-gray-700/30"
                  }`}
                >
                  <span
                    className={`text-sm ${isCurrentUser ? "text-white font-semibold" : "text-gray-300"}`}
                  >
                    {player.username}
                    {isCurrentUser && (
                      <span className="ml-1 text-xs text-gray-500">(you)</span>
                    )}
                    {isPlayerWinner && (
                      <span className="ml-1 text-xs text-yellow-400">
                        &#9733;
                      </span>
                    )}
                  </span>
                  <span className="text-sm text-white font-mono tabular-nums">
                    {player.score} pts
                  </span>
                </div>
              );
            })}
        </div>
        {winnerPlayer && (
          <p className="mt-2 text-xs text-gray-500">
            Winner: {winnerPlayer.username}
          </p>
        )}
        <Link
          to="/replay/$gameID"
          params={{ gameID: String(game.game_id) }}
        >
          <Button
            variant="outline"
            size="sm"
            className="w-full mt-3 border-gray-600 text-gray-300 hover:text-white hover:bg-gray-700"
          >
            <Play className="h-4 w-4 mr-2" />
            Watch Replay
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
}

function HistoryPage() {
  const auth = useAuth();
  const {
    data: games,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["gameHistory"],
    queryFn: fetchGameHistory,
  });

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-3xl mx-auto space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-white">Game History</h1>
          <p className="text-gray-400 mt-1">
            Your past games and scores
          </p>
        </div>

        {isLoading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading game history...</p>
          </div>
        ) : isError ? (
          <div className="text-center py-12">
            <p className="text-red-400">
              Failed to load game history. Please try again.
            </p>
          </div>
        ) : !games || games.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-400">
              No completed games yet. Play some games to see your history!
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {games.map((game) => (
              <GameCard
                key={game.game_id}
                game={game}
                currentUserID={auth.user?.userID ?? 0}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
