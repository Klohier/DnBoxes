import { createFileRoute } from "@tanstack/react-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Users,
  Zap,
  MessageSquare,
  Sparkles,
  Trophy,
  Heart,
  BookOpen,
} from "lucide-react";

export const Route = createFileRoute("/about")({
  component: About,
  head: () => ({
    meta: [
      {
        title: "About Dots & Boxes - How to Play & Game Rules",
      },
      {
        name: "description",
        content:
          "Learn how to play Dots & Boxes online. Discover game rules, strategy tips, and features. Play the classic pencil-and-paper game with friends or AI opponents.",
      },
      {
        name: "keywords",
        content:
          "dots and boxes rules, how to play dots and boxes, dots and boxes strategy, dots and boxes tips, game rules, online multiplayer",
      },
      // Open Graph
      {
        property: "og:title",
        content: "About Dots & Boxes - How to Play & Game Rules",
      },
      {
        property: "og:description",
        content:
          "Learn how to play Dots & Boxes online. Discover game rules, winning strategies, and multiplayer features.",
      },
      {
        property: "og:type",
        content: "website",
      },
      {
        property: "og:url",
        content: "https://dotsandboxesonline.com/about",
      },

      // Twitter Card
      {
        name: "twitter:card",
        content: "summary_large_image",
      },
      {
        name: "twitter:title",
        content: "About Dots & Boxes - How to Play",
      },
      {
        name: "twitter:description",
        content: "Learn the rules and strategies of this classic game.",
      },
    ],
    links: [
      {
        rel: "canonical",
        href: "https://dotsandboxesonline.com/about",
      },
    ],
    scripts: [
      {
        type: "application/ld+json",
        children: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "HowTo",
          name: "How to Play Dots & Boxes",
          description:
            "Learn how to play the classic Dots & Boxes game online with friends or AI opponents.",
          step: [
            {
              "@type": "HowToStep",
              name: "Draw a Line",
              text: "Take turns connecting two adjacent points with a horizontal or vertical line.",
            },
            {
              "@type": "HowToStep",
              name: "Complete a Box",
              text: "Close the fourth side of a box to claim it and earn another turn.",
            },
            {
              "@type": "HowToStep",
              name: "Win the Game",
              text: "The player who captures the most boxes when the grid is full wins.",
            },
          ],
        }),
      },
      {
        type: "application/ld+json",
        children: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "FAQPage",
          mainEntity: [
            {
              "@type": "Question",
              name: "How do you play Dots and Boxes?",
              acceptedAnswer: {
                "@type": "Answer",
                text: "Players take turns drawing a line between two adjacent points (horizontal or vertical). When a player completes the fourth side of a box, they claim it and take another turn. The game ends when all boxes are claimed, and the player with the most boxes wins.",
              },
            },

            {
              "@type": "Question",
              name: "What is a good strategy for Dots and Boxes?",
              acceptedAnswer: {
                "@type": "Answer",
                text: "In early game, avoid giving your opponent easy boxes. In mid game, look for chain opportunities. In late game, control the number of available boxes and force your opponent into unfavorable positions.",
              },
            },
          ],
        }),
      },
    ],
  }),
});

function About() {
  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Hero Section */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader className="text-center">
            <CardTitle className="text-4xl font-bold text-white mb-2">
              About Dots & Boxes
            </CardTitle>
            <CardDescription className="text-lg text-gray-300">
              A modern take on the timeless pencil-and-paper classic
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-gray-300 text-center">
              Welcome to{" "}
              <span className="font-semibold text-white">
                Dots & Boxes Online
              </span>
              , where players from around the world come together to enjoy this
              strategic game. Play against friends, challenge AI opponents, or
              compete in real-time multiplayer matches.
            </p>
          </CardContent>
        </Card>

        {/* How to Play */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <CardTitle className="text-2xl text-white flex items-center gap-2">
              <BookOpen className="h-6 w-6 text-cyan-400" />
              How to Play Dots and Boxes
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold text-white mb-2">
                The Basics
              </h3>
              <p className="text-gray-300">
                Dots & Boxes is played on a grid. Players take turns drawing a
                line between two adjacent points (horizontal or vertical). When
                a player completes the fourth side of a box, they claim it and
                take another turn.
              </p>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-white mb-2">
                Winning the Game
              </h3>
              <p className="text-gray-300">
                The game ends when all boxes are claimed. The player who has
                captured the most boxes wins! In case of a tie, the game is a
                draw.
              </p>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-white mb-2">
                Strategy Tips
              </h3>
              <ul className="text-gray-300 space-y-2 list-disc list-inside">
                <li>
                  Early game: Try to avoid giving your opponent easy boxes by
                  not completing three sides of any box
                </li>
                <li>
                  Mid game: Look for chain opportunities where capturing one box
                  leads to capturing many
                </li>
                <li>
                  Late game: Control the number of boxes available and force
                  your opponent into unfavorable positions
                </li>
              </ul>
            </div>
          </CardContent>
        </Card>

        {/* Why We Built This */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <CardTitle className="text-2xl text-white flex items-center gap-2">
              <Heart className="h-6 w-6 text-red-400" />
              Why We Built This
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-gray-300">
              We believe games are more than entertainment—they're about
              connection, strategy, and shared moments. Dots & Boxes is easy to
              learn but endlessly challenging, making it perfect for quick
              casual matches or intense competitive play.
            </p>
          </CardContent>
        </Card>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card className="bg-gray-800 border-gray-700">
            <CardHeader>
              <CardTitle className="text-xl text-white flex items-center gap-2">
                <Users className="h-5 w-5 text-blue-400" />
                Real-Time Multiplayer
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-300 text-sm">
                Challenge friends or match with players globally. Create private
                lobbies or join public games.
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-800 border-gray-700">
            <CardHeader>
              <CardTitle className="text-xl text-white flex items-center gap-2">
                <Zap className="h-5 w-5 text-yellow-400" />
                Bot Practice Mode
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-300 text-sm">
                Hone your skills against AI opponents before taking on real
                players.
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-800 border-gray-700">
            <CardHeader>
              <CardTitle className="text-xl text-white flex items-center gap-2">
                <MessageSquare className="h-5 w-5 text-green-400" />
                Live Chat
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-300 text-sm">
                Communicate with opponents during games or in the global lobby.
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-800 border-gray-700">
            <CardHeader>
              <CardTitle className="text-xl text-white flex items-center gap-2">
                <Sparkles className="h-5 w-5 text-purple-400" />
                Custom Game Options
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-300 text-sm">
                Choose your board size and number of players to customize your
                experience.
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Mission Statement */}
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <CardTitle className="text-2xl text-white flex items-center gap-2">
              <Trophy className="h-6 w-6 text-yellow-500" />
              The Mission
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-gray-300">
              The mission is to keep the spirit of Dots & Boxes alive in the
              digital age. Whether you're a first-time player or a seasoned
              strategist, you'll find a welcoming community here to play, learn,
              and connect.
            </p>
            <p className="text-center font-medium text-lg text-white">
              Join us, draw your first line, and see where the boxes take you.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
