import { createFileRoute, Link } from "@tanstack/react-router";
import { useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
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
});

function DotsGrid() {
  const dots: { x: number; y: number }[] = [];
  const cols = 4;
  const rows = 4;
  const spacing = 60;
  const offsetX = 30;
  const offsetY = 30;

  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      dots.push({ x: offsetX + c * spacing, y: offsetY + r * spacing });
    }
  }

  const lines = [
    // Some horizontal lines
    { x1: 30, y1: 30, x2: 90, y2: 30, color: "#60a5fa", delay: "0s" },
    { x1: 90, y1: 30, x2: 150, y2: 30, color: "#60a5fa", delay: "0.3s" },
    { x1: 30, y1: 90, x2: 90, y2: 90, color: "#f472b6", delay: "0.6s" },
    { x1: 150, y1: 90, x2: 210, y2: 90, color: "#60a5fa", delay: "0.9s" },
    { x1: 90, y1: 150, x2: 150, y2: 150, color: "#f472b6", delay: "1.2s" },
    { x1: 150, y1: 150, x2: 210, y2: 150, color: "#f472b6", delay: "1.5s" },
    { x1: 30, y1: 210, x2: 90, y2: 210, color: "#60a5fa", delay: "1.8s" },
    // Some vertical lines
    { x1: 30, y1: 30, x2: 30, y2: 90, color: "#f472b6", delay: "0.4s" },
    { x1: 90, y1: 30, x2: 90, y2: 90, color: "#60a5fa", delay: "0.7s" },
    { x1: 210, y1: 30, x2: 210, y2: 90, color: "#f472b6", delay: "1.0s" },
    { x1: 150, y1: 90, x2: 150, y2: 150, color: "#60a5fa", delay: "1.3s" },
    { x1: 210, y1: 90, x2: 210, y2: 150, color: "#60a5fa", delay: "1.6s" },
    { x1: 30, y1: 150, x2: 30, y2: 210, color: "#f472b6", delay: "1.9s" },
  ];

  // Completed box fill
  const completedBoxes = [
    {
      x: 31,
      y: 31,
      color: "rgba(96, 165, 250, 0.15)",
      delay: "1.1s",
    },
  ];

  return (
    <svg
      viewBox="0 0 240 240"
      className="w-56 h-56 md:w-72 md:h-72"
      aria-hidden="true"
    >
      {/* Completed box backgrounds */}
      {completedBoxes.map((box, i) => (
        <rect
          key={`box-${i}`}
          x={box.x}
          y={box.y}
          width={58}
          height={58}
          fill={box.color}
          className="animate-fade-in"
          style={{ animationDelay: box.delay }}
        />
      ))}

      {/* Lines */}
      {lines.map((line, i) => (
        <line
          key={`line-${i}`}
          x1={line.x1}
          y1={line.y1}
          x2={line.x2}
          y2={line.y2}
          stroke={line.color}
          strokeWidth={3}
          strokeLinecap="round"
          className="animate-draw-line"
          style={{ animationDelay: line.delay }}
        />
      ))}

      {/* Dots */}
      {dots.map((dot, i) => (
        <circle
          key={`dot-${i}`}
          cx={dot.x}
          cy={dot.y}
          r={5}
          className="fill-white animate-pulse-dot"
          style={{ animationDelay: `${i * 0.1}s` }}
        />
      ))}
    </svg>
  );
}

const features = [
  {
    icon: Users,
    title: "Multiplayer",
    description: "Play with friends or match with players worldwide in real time.",
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

function Index() {
  const auth = useAuth();

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Hero Section */}
      <section className="relative overflow-hidden">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-b from-blue-900/20 via-gray-900 to-gray-900" />
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-blue-500/5 rounded-full blur-3xl" />

        <div className="relative max-w-5xl mx-auto px-4 pt-16 pb-20 md:pt-24 md:pb-28">
          <div className="flex flex-col md:flex-row items-center gap-10 md:gap-16">
            {/* Text content */}
            <div className="flex-1 text-center md:text-left space-y-6">
              <h1 className="text-5xl md:text-6xl font-extrabold text-white tracking-tight leading-tight">
                Dots &{" "}
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-cyan-300">
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

            {/* Animated dots grid illustration */}
            <div className="flex-shrink-0">
              <DotsGrid />
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
                desc: "Take turns connecting two adjacent dots with a horizontal or vertical line.",
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
        <h2 className="text-3xl font-bold text-white mb-4">
          Ready to play?
        </h2>
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
