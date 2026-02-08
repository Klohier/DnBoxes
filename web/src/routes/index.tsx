import { createFileRoute, Link } from "@tanstack/react-router";
import { useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import Grid from "@/components/Grid";
import { Box as BoxType } from "@/types/websocket";
import { useState, useEffect, useCallback } from "react";
import {
  Users,
  Zap,
  MessageSquare,
  Sparkles,
  Trophy,
  ArrowRight,
} from "lucide-react";

export const Route = createFileRoute("/")({
  component: Index,
  head: () => ({
    meta: [
      {
        title: "Dots & Boxes - Play the Classic Game Online",
      },
      {
        name: "description",
        content:
          "Play Dots & Boxes online! Challenge friends, battle AI opponents, or climb the leaderboard in this classic pencil-and-paper game reimagined for the web.",
      },
      {
        name: "keywords",
        content:
          "dots and boxes, online game, multiplayer game, strategy game, board game, classic game",
      },
      // Open Graph
      {
        property: "og:title",
        content: "Dots & Boxes - Play the Classic Game Online",
      },
      {
        property: "og:description",
        content:
          "Challenge friends, battle bots, or climb the leaderboard in this classic pencil-and-paper game.",
      },
      {
        property: "og:type",
        content: "website",
      },
      {
        property: "og:url",
        content: "https://dotsandboxesonline.com/",
      },

      {
        property: "og:site_name",
        content: "Dots & Boxes Online",
      },
      // Twitter Card
      {
        name: "twitter:card",
        content: "summary_large_image",
      },
      {
        name: "twitter:title",
        content: "Dots & Boxes - Play the Classic Game Online",
      },
      {
        name: "twitter:description",
        content: "Challenge friends, battle bots, or climb the leaderboard.",
      },
    ],
    links: [
      {
        rel: "canonical",
        href: "https://dotsandboxesonline.com/",
      },
    ],
    scripts: [
      {
        type: "application/ld+json",
        children: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "WebApplication",
          name: "Dots & Boxes Online",
          description: "Play the classic Dots & Boxes game online",
          url: "https://dotsandboxesonline.com",
          applicationCategory: "Game",
          operatingSystem: "Web Browser",
        }),
      },
    ],
  }),
});

// --- Demo grid that replays a short game using the real Grid component ---

const DEMO_BOARD_SIZE = 3;
const DEMO_COLORS: Record<number, string> = { 1: "#60a5fa", 2: "#f472b6" };
const DEMO_TURN_MAP: Record<number, number> = { 0: 1, 1: 2 };

type EdgeName = "top_edge" | "right_edge" | "bottom_edge" | "left_edge";

interface Move {
  row: number;
  col: number;
  edge: EdgeName;
  player: number; // turn_order: 0 or 1
}

function createEmptyGrid(size: number): BoxType[] {
  const boxes: BoxType[] = [];
  for (let r = 0; r < size; r++) {
    for (let c = 0; c < size; c++) {
      boxes.push({
        row: r,
        col: c,
        top_edge: false,
        right_edge: false,
        bottom_edge: false,
        left_edge: false,
        owner_turn: null,
      });
    }
  }
  return boxes;
}

function getSharedEdge(
  row: number,
  col: number,
  edge: EdgeName,
  size: number,
): { row: number; col: number; edge: EdgeName } | null {
  switch (edge) {
    case "top_edge":
      return row > 0 ? { row: row - 1, col, edge: "bottom_edge" } : null;
    case "bottom_edge":
      return row < size - 1 ? { row: row + 1, col, edge: "top_edge" } : null;
    case "left_edge":
      return col > 0 ? { row, col: col - 1, edge: "right_edge" } : null;
    case "right_edge":
      return col < size - 1 ? { row, col: col + 1, edge: "left_edge" } : null;
  }
}

function isBoxComplete(box: BoxType): boolean {
  return box.top_edge && box.right_edge && box.bottom_edge && box.left_edge;
}

function applyMove(boxes: BoxType[], move: Move): BoxType[] {
  const newBoxes = boxes.map((b) => ({ ...b }));
  const size = DEMO_BOARD_SIZE;

  const idx = move.row * size + move.col;
  newBoxes[idx][move.edge] = true;

  const shared = getSharedEdge(move.row, move.col, move.edge, size);
  if (shared) {
    const sharedIdx = shared.row * size + shared.col;
    newBoxes[sharedIdx][shared.edge] = true;
  }

  for (const box of newBoxes) {
    if (box.owner_turn === null && isBoxComplete(box)) {
      box.owner_turn = move.player;
    }
  }

  return newBoxes;
}

// A short scripted game — move 7 completes box (0,0) with the real boxPop animation
const DEMO_MOVES: Move[] = [
  { row: 0, col: 0, edge: "top_edge", player: 0 },
  { row: 1, col: 1, edge: "right_edge", player: 1 },
  { row: 0, col: 0, edge: "left_edge", player: 0 },
  { row: 2, col: 0, edge: "bottom_edge", player: 1 },
  { row: 0, col: 0, edge: "right_edge", player: 0 },
  { row: 1, col: 2, edge: "top_edge", player: 1 },
  { row: 0, col: 0, edge: "bottom_edge", player: 0 },
  { row: 0, col: 1, edge: "top_edge", player: 0 },
  { row: 2, col: 2, edge: "right_edge", player: 1 },
  { row: 1, col: 0, edge: "left_edge", player: 0 },
  { row: 2, col: 1, edge: "bottom_edge", player: 1 },
  { row: 1, col: 0, edge: "bottom_edge", player: 0 },
];

function DemoGrid() {
  const [boxes, setBoxes] = useState(() => createEmptyGrid(DEMO_BOARD_SIZE));
  const [moveIndex, setMoveIndex] = useState(0);

  const reset = useCallback(() => {
    setBoxes(createEmptyGrid(DEMO_BOARD_SIZE));
    setMoveIndex(0);
  }, []);

  useEffect(() => {
    if (moveIndex >= DEMO_MOVES.length) {
      const timeout = setTimeout(reset, 3000);
      return () => clearTimeout(timeout);
    }

    const timeout = setTimeout(
      () => {
        setBoxes((prev) => applyMove(prev, DEMO_MOVES[moveIndex]));
        setMoveIndex((i) => i + 1);
      },
      moveIndex === 0 ? 600 : 800,
    );

    return () => clearTimeout(timeout);
  }, [moveIndex, reset]);

  return (
    <div className="w-56 h-56 md:w-72 md:h-72 pointer-events-none">
      <Grid
        gameID={0}
        boxes={boxes}
        userColors={DEMO_COLORS}
        boardSize={DEMO_BOARD_SIZE}
        userID={0}
        handleClick={() => {}}
        turnToUserIdMap={DEMO_TURN_MAP}
      />
    </div>
  );
}

// --- Feature cards ---

const features = [
  {
    icon: Users,
    title: "Multiplayer",
    description:
      "Play with friends or match with players worldwide in real time.",
    color: "text-blue-400",
    bg: "bg-blue-400/10",
  },
  {
    icon: Zap,
    title: "Bot Practice",
    description: "Sharpen your strategy against AI opponents at your own pace.",
    color: "text-yellow-400",
    bg: "bg-yellow-400/10",
  },
  {
    icon: MessageSquare,
    title: "Live Chat",
    description: "Talk with opponents during matches or in the global lobby.",
    color: "text-green-400",
    bg: "bg-green-400/10",
  },
  {
    icon: Trophy,
    title: "Leaderboard",
    description: "Climb the ranks and see how you stack up against the best.",
    color: "text-purple-400",
    bg: "bg-purple-400/10",
  },
  {
    icon: Sparkles,
    title: "Custom Games",
    description: "Choose board sizes and player counts for your ideal match.",
    color: "text-cyan-400",
    bg: "bg-cyan-400/10",
  },
];

// --- Page ---

function Index() {
  const auth = useAuth();

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Hero Section */}
      <section className="relative overflow-hidden">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-linear-to-b from-blue-900/20 via-gray-900 to-gray-900" />
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-200 h-100 bg-blue-500/5 rounded-full blur-3xl" />

        <div className="relative max-w-5xl mx-auto px-4 pt-16 pb-20 md:pt-24 md:pb-28">
          <div className="flex flex-col md:flex-row items-center gap-10 md:gap-16">
            {/* Text content */}
            <div className="flex-1 text-center md:text-left space-y-6">
              <h1 className="text-5xl md:text-6xl font-extrabold text-white tracking-tight leading-tight">
                Dots &{" "}
                <span className="text-transparent bg-clip-text bg-linear-to-r from-blue-400 to-cyan-300">
                  Boxes
                </span>
              </h1>
              <p className="text-xl text-gray-300 max-w-lg">
                The classic pencil-and-paper game, reimagined for the web.
                Challenge friends, battle bots, or climb the leaderboard.
              </p>
              <div className="flex flex-col sm:flex-row gap-3 justify-center md:justify-start">
                {auth.isAuthenticated ? (
                  <Link to="/play">
                    <Button
                      size="lg"
                      className="w-full sm:w-auto bg-blue-600 hover:bg-blue-500 text-white text-lg px-8 py-3 h-auto cursor-pointer"
                    >
                      Play Now
                      <ArrowRight className="ml-1 h-5 w-5" />
                    </Button>
                  </Link>
                ) : (
                  <>
                    <Link to="/register">
                      <Button
                        size="lg"
                        className="w-full sm:w-auto bg-blue-600 hover:bg-blue-500 text-white text-lg px-8 py-3 h-auto cursor-pointer"
                      >
                        Get Started
                        <ArrowRight className="ml-1 h-5 w-5" />
                      </Button>
                    </Link>
                    <Link to="/login">
                      <Button
                        size="lg"
                        variant="outline"
                        className="w-full sm:w-auto border-gray-600 text-gray-200 hover:bg-gray-800 hover:text-white text-lg px-8 py-3 h-auto cursor-pointer"
                      >
                        Sign In
                      </Button>
                    </Link>
                  </>
                )}
              </div>
            </div>

            {/* Live demo of the actual game grid */}
            <div className="shrink-0">
              <DemoGrid />
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="max-w-5xl mx-auto px-4 pb-20">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="group rounded-xl border border-gray-700/50 bg-gray-800/50 p-6 hover:border-gray-600 hover:bg-gray-800 transition-colors"
            >
              <div
                className={`inline-flex items-center justify-center w-10 h-10 rounded-lg ${feature.bg} mb-4`}
              >
                <feature.icon className={`h-5 w-5 ${feature.color}`} />
              </div>
              <h3 className="text-lg font-semibold text-white mb-1">
                {feature.title}
              </h3>
              <p className="text-sm text-gray-400">{feature.description}</p>
            </div>
          ))}
        </div>
      </section>

      {/* How It Works */}
      <section className="border-t border-gray-800 bg-gray-900/50">
        <div className="max-w-5xl mx-auto px-4 py-16">
          <h2 className="text-2xl font-bold text-white text-center mb-10">
            How It Works
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              {
                step: "1",
                title: "Draw a Line",
                desc: "Take turns connecting two adjacent points with a horizontal or vertical line.",
              },
              {
                step: "2",
                title: "Complete a Box",
                desc: "Close the fourth side of a box to claim it and earn another turn.",
              },
              {
                step: "3",
                title: "Win the Game",
                desc: "The player who captures the most boxes when the grid is full wins.",
              },
            ].map((item) => (
              <div key={item.step} className="text-center space-y-3">
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-blue-600/20 text-blue-400 text-xl font-bold">
                  {item.step}
                </div>
                <h3 className="text-lg font-semibold text-white">
                  {item.title}
                </h3>
                <p className="text-sm text-gray-400">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Bottom CTA */}
      <section className="max-w-5xl mx-auto px-4 py-16 text-center">
        <h2 className="text-3xl font-bold text-white mb-4">Ready to play?</h2>
        <p className="text-gray-400 mb-8 max-w-md mx-auto">
          Jump into a game in seconds. No download required.
        </p>
        {auth.isAuthenticated ? (
          <Link to="/play">
            <Button
              size="lg"
              className="bg-blue-600 hover:bg-blue-500 text-white text-lg px-10 py-3 h-auto cursor-pointer"
            >
              Start Playing
              <ArrowRight className="ml-1 h-5 w-5" />
            </Button>
          </Link>
        ) : (
          <Link to="/register">
            <Button
              size="lg"
              className="bg-blue-600 hover:bg-blue-500 text-white text-lg px-10 py-3 h-auto cursor-pointer"
            >
              Create Free Account
              <ArrowRight className="ml-1 h-5 w-5" />
            </Button>
          </Link>
        )}
      </section>
    </div>
  );
}
