import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { fetchLeaderboard, LeaderboardEntry } from "@/api/fetchStats";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export const Route = createFileRoute("/leaderboard")({
  component: LeaderboardPage,
});

function RankBadge({ rank }: { rank: number }) {
  if (rank === 1)
    return (
      <span className="text-lg font-bold text-yellow-400 w-8 text-center">
        1st
      </span>
    );
  if (rank === 2)
    return (
      <span className="text-lg font-bold text-gray-300 w-8 text-center">
        2nd
      </span>
    );
  if (rank === 3)
    return (
      <span className="text-lg font-bold text-amber-600 w-8 text-center">
        3rd
      </span>
    );
  return (
    <span className="text-sm text-gray-500 w-8 text-center">{rank}</span>
  );
}

function LeaderboardTable({
  entries,
  valueLabel,
}: {
  entries: LeaderboardEntry[] | null;
  valueLabel: string;
}) {
  if (!entries || entries.length === 0) {
    return (
      <p className="text-gray-500 text-sm py-4 text-center">
        No data yet. Play some games!
      </p>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between px-3 py-1 text-xs text-gray-500 uppercase tracking-wider">
        <div className="flex items-center gap-3">
          <span className="w-8">Rank</span>
          <span>Player</span>
        </div>
        <span>{valueLabel}</span>
      </div>
      {entries.map((entry) => (
        <div
          key={entry.userID}
          className={`flex items-center justify-between px-3 py-3 rounded-lg ${
            entry.rank <= 3
              ? "bg-gray-700/50 border border-gray-600"
              : "bg-gray-800/50"
          }`}
        >
          <div className="flex items-center gap-3">
            <RankBadge rank={entry.rank} />
            <span className="text-white font-medium">{entry.username}</span>
          </div>
          <span className="text-white font-semibold tabular-nums">
            {entry.value.toLocaleString()}
          </span>
        </div>
      ))}
    </div>
  );
}

function LeaderboardPage() {
  const {
    data: leaderboard,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["leaderboard"],
    queryFn: fetchLeaderboard,
  });

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-white">Leaderboard</h1>
          <p className="text-gray-400 mt-1">
            Top players ranked by performance
          </p>
        </div>

        {isLoading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading leaderboard...</p>
          </div>
        ) : isError ? (
          <div className="text-center py-12">
            <p className="text-red-400">
              Failed to load leaderboard. Please try again.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card className="bg-gray-800 border-gray-700">
              <CardHeader>
                <CardTitle className="text-xl text-white flex items-center gap-2">
                  <span className="text-yellow-400">&#9733;</span>
                  Most Wins
                </CardTitle>
              </CardHeader>
              <CardContent>
                <LeaderboardTable
                  entries={leaderboard?.mostWins ?? null}
                  valueLabel="Wins"
                />
              </CardContent>
            </Card>

            <Card className="bg-gray-800 border-gray-700">
              <CardHeader>
                <CardTitle className="text-xl text-white flex items-center gap-2">
                  <span className="text-blue-400">&#9632;</span>
                  Most Boxes Completed
                </CardTitle>
              </CardHeader>
              <CardContent>
                <LeaderboardTable
                  entries={leaderboard?.mostBoxes ?? null}
                  valueLabel="Boxes"
                />
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}
