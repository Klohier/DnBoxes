import { useEffect, useState } from "react";
import { useWebSocket } from "@/WebSocketContext";
import { LobbyList } from "@/components/Lobby";
import { LobbyModal } from "@/components/LobbyModal";
import { BotGameModal } from "@/components/BotGameModal";
import Chatbox from "../../components/ChatBox";
import { Button } from "@/components/ui/button";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import { CreateLobbyData, Lobby } from "../../types/lobby";
import { fetchLobbies, createLobby } from "@/api/lobby";
import { fetchMyStats } from "@/api/fetchStats";
import { useAuth } from "@/AuthContext";

import { toast } from "sonner";
import axios from "axios";

export const Route = createFileRoute("/_authenticated/")({
  component: Index,
});

interface Game {
  game_id: number;
}

interface CreateBotGameData {
  board_size: number;
  num_bots: number;
}

function Index() {
  const navigate = useNavigate();
  const auth = useAuth();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isBotModalOpen, setIsBotModalOpen] = useState(false);
  const [showUpgrade, setShowUpgrade] = useState(false);
  const [upgradeUsername, setUpgradeUsername] = useState("");
  const [upgradePassword, setUpgradePassword] = useState("");
  const [upgradeError, setUpgradeError] = useState("");
  const queryClient = useQueryClient();
  const { subscribe } = useWebSocket();

  // Fetch lobbies
  const { data: lobbies = [], isLoading } = useQuery<Lobby[]>({
    queryKey: ["lobbies"],
    queryFn: fetchLobbies,
  });

  // Fetch user stats
  const { data: stats } = useQuery({
    queryKey: ["myStats"],
    queryFn: fetchMyStats,
  });

  // Create lobby mutation
  const createLobbyMutation = useMutation({
    mutationFn: createLobby,
    onSuccess: (newLobby) => {
      queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) => [
        ...old,
        newLobby,
      ]);
      setIsModalOpen(false);
      void navigate({
        to: "/lobby/$lobbyID",
        params: { lobbyID: newLobby.lobby_id },
      });
    },
    onError: (error) => {
      console.error("Failed to create lobby:", error);
      toast.error("Failed to create lobby");
    },
  });

  // Create bot game mutation
  const createBotGameMutation = useMutation({
    mutationFn: async (data: CreateBotGameData) => {
      const response = await axios.post<Game>(
        `/api/v1/games/create-bot-game`,
        {
          human_player_id: auth.user?.userID,
          board_size: data.board_size,
          num_bots: data.num_bots,
        },
        {
          headers: { "Content-Type": "application/json" },
        },
      );
      return response.data;
    },
    onSuccess: (game) => {
      toast.success("Bot game created!");
      setIsBotModalOpen(false);
      void navigate({
        to: "/game/$gameID",
        params: { gameID: String(game.game_id) },
      });
    },
    onError: (error) => {
      console.error("Failed to create bot game:", error);
      toast.error("Failed to create bot game");
    },
  });

  const handleCreateLobby = async (values: CreateLobbyData): Promise<void> => {
    await createLobbyMutation.mutateAsync(values);
  };

  const handleCreateBotGame = async (
    values: CreateBotGameData,
  ): Promise<void> => {
    if (!auth.isAuthenticated) {
      toast.error("Please login to play");
      return;
    }
    await createBotGameMutation.mutateAsync(values);
  };

  useEffect(() => {
    const unsubscribe = subscribe((msg) => {
      if (msg.type === "lobby_created") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) => [
          ...old,
          msg.payload,
        ]);
      } else if (msg.type === "lobby_deleted") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
          old.filter((l) => l.lobby_id !== msg.payload.lobby_id),
        );
      } else if (msg.type === "lobby_updated") {
        queryClient.setQueryData<Lobby[]>(["lobbies"], (old = []) =>
          old.map((l) =>
            l.lobby_id === msg.payload.lobby_id ? { ...l, ...msg.payload } : l,
          ),
        );
      }
    });

    return unsubscribe;
  }, [subscribe, queryClient]);

  return (
    <>
      <div className="min-h-screen bg-gray-900 p-4">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {/* Left Section - Lobby Controls & List */}
            <div className="lg:col-span-2 flex flex-col gap-4">
              <div className="bg-gray-800 rounded-lg p-6 flex-shrink-0">
                <h1 className="text-3xl font-bold text-white mb-4">
                  Dots & Boxes
                </h1>
                <div className="flex gap-3">
                  <Button
                    onClick={() => setIsModalOpen(true)}
                    className="flex-1"
                  >
                    + Create Lobby
                  </Button>
                  <Button
                    onClick={() => setIsBotModalOpen(true)}
                    className="flex-1"
                    variant="secondary"
                  >
                    + Create Bot Match
                  </Button>
                </div>
              </div>

              <div className="bg-gray-800 rounded-lg overflow-hidden flex flex-col">
                <div className="p-6 pb-4 border-b border-gray-700 flex-shrink-0">
                  <h2 className="text-xl font-bold text-white">
                    Available Lobbies
                  </h2>
                </div>
                <div className="overflow-y-auto p-6 pt-4 h-[400px]">
                  {isLoading ? (
                    <p className="text-gray-400">Loading lobbies...</p>
                  ) : (
                    <LobbyList lobbies={lobbies} />
                  )}
                </div>
              </div>
            </div>

            {/* Right Section - Profile & Chat */}
            <div className="lg:col-span-1 flex flex-col gap-4">
              <div className="bg-gray-800 rounded-lg p-6 flex-shrink-0">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-xl font-bold text-white">Profile</h2>
                  <span className="text-sm text-gray-400">
                    {auth.user?.username}
                    {auth.user?.isGuest && (
                      <span className="ml-2 text-xs bg-yellow-600 text-yellow-100 px-1.5 py-0.5 rounded">
                        Guest
                      </span>
                    )}
                  </span>
                </div>

                {auth.user?.isGuest && !showUpgrade && (
                  <div className="mb-4 p-3 bg-yellow-900/30 border border-yellow-700 rounded-lg">
                    <p className="text-yellow-200 text-sm mb-2">
                      Create an account to save your stats and appear on the leaderboard.
                    </p>
                    <Button
                      size="sm"
                      onClick={() => setShowUpgrade(true)}
                    >
                      Upgrade Account
                    </Button>
                  </div>
                )}

                {auth.user?.isGuest && showUpgrade && (
                  <div className="mb-4 p-3 bg-gray-700 rounded-lg">
                    <h3 className="text-white text-sm font-semibold mb-2">
                      Create Your Account
                    </h3>
                    <form
                      onSubmit={async (e) => {
                        e.preventDefault();
                        setUpgradeError("");
                        try {
                          await auth.upgradeAccount({
                            username: upgradeUsername,
                            password: upgradePassword,
                          });
                          toast.success("Account upgraded successfully!");
                          setShowUpgrade(false);
                        } catch (err) {
                          if (axios.isAxiosError(err)) {
                            setUpgradeError(
                              err.response?.data?.message || "Upgrade failed",
                            );
                          } else {
                            setUpgradeError("Upgrade failed");
                          }
                        }
                      }}
                      className="space-y-2"
                    >
                      <input
                        type="text"
                        placeholder="Choose a username"
                        value={upgradeUsername}
                        onChange={(e) => setUpgradeUsername(e.target.value)}
                        required
                        className="w-full px-3 py-1.5 bg-gray-600 text-white rounded text-sm placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                      <input
                        type="password"
                        placeholder="Choose a password"
                        value={upgradePassword}
                        onChange={(e) => setUpgradePassword(e.target.value)}
                        required
                        className="w-full px-3 py-1.5 bg-gray-600 text-white rounded text-sm placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                      {upgradeError && (
                        <p className="text-red-400 text-xs">{upgradeError}</p>
                      )}
                      <div className="flex gap-2">
                        <Button type="submit" size="sm" className="flex-1">
                          Save
                        </Button>
                        <Button
                          type="button"
                          size="sm"
                          variant="secondary"
                          onClick={() => {
                            setShowUpgrade(false);
                            setUpgradeError("");
                          }}
                        >
                          Cancel
                        </Button>
                      </div>
                    </form>
                  </div>
                )}

                <div className="space-y-3 text-sm">
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Games Played</span>
                    <span className="text-white font-semibold">
                      {stats?.gamesPlayed ?? 0}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Wins</span>
                    <span className="text-green-400 font-semibold">
                      {stats?.wins ?? 0}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Losses</span>
                    <span className="text-red-400 font-semibold">
                      {stats?.losses ?? 0}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Win Rate</span>
                    <span className="text-white font-semibold">
                      {stats?.winRate !== undefined
                        ? `${stats.winRate.toFixed(1)}%`
                        : "0%"}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Boxes Completed</span>
                    <span className="text-blue-400 font-semibold">
                      {stats?.totalBoxes ?? 0}
                    </span>
                  </div>
                </div>
                <div className="mt-4 pt-3 border-t border-gray-700 flex gap-4">
                  <Link
                    to="/leaderboard"
                    className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
                  >
                    View Leaderboard
                  </Link>
                  <Link
                    to="/history"
                    className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
                  >
                    Game History
                  </Link>
                </div>
              </div>

              <div className="bg-gray-800 rounded-lg overflow-hidden flex flex-col">
                <div className="p-4 border-b border-gray-700 flex-shrink-0">
                  <h2 className="text-lg font-bold text-white">Global Chat</h2>
                </div>

                <Chatbox topic="chat:global" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <LobbyModal
        open={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateLobby}
      />

      <BotGameModal
        open={isBotModalOpen}
        onClose={() => setIsBotModalOpen(false)}
        onSubmit={handleCreateBotGame}
      />
    </>
  );
}
