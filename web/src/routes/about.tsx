import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/about")({
  component: About,
});

function About() {
  return (
    <div className="p-6 max-w-3xl mx-auto space-y-6">
      <h1 className="text-3xl font-bold text-center">About Us</h1>

      <p className="text-lg text-gray-700">
        Welcome to <span className="font-semibold">Dots & Boxes Online</span>,
        the home for players around the world who love this timeless
        pencil-and-paper classic. Our platform brings the game to life in a
        modern, multiplayer experience you can enjoy anywhere.
      </p>

      <section className="space-y-3">
        <h2 className="text-2xl font-semibold">Why We Built This</h2>
        <p className="text-gray-700">
          We believe games are more than entertainment—they’re about connection,
          strategy, and shared moments. Dots & Boxes is easy to learn but
          endlessly challenging, making it perfect for quick casual matches or
          intense competitive play.
        </p>
      </section>

      <section className="space-y-3">
        <h2 className="text-2xl font-semibold">What You’ll Find Here</h2>
        <ul className="list-disc list-inside text-gray-700 space-y-1">
          <li>
            <span className="font-medium">Real-Time Multiplayer</span> –
            Challenge friends or match with players globally.
          </li>
          <li>
            <span className="font-medium">Smart Matchmaking</span> – Play at
            your level and climb the leaderboards.
          </li>
          <li>
            <span className="font-medium">Chat & Community</span> – Celebrate
            wins and share strategies.
          </li>
          <li>
            <span className="font-medium">Custom Game Modes</span> – Classic
            rules or new twists to keep things fresh.
          </li>
        </ul>
      </section>

      <section className="space-y-3">
        <h2 className="text-2xl font-semibold">Our Mission</h2>
        <p className="text-gray-700">
          Our mission is to keep the spirit of Dots & Boxes alive in the digital
          age. Whether you’re a first-time player or a seasoned strategist,
          you’ll find a place here to play, learn, and connect.
        </p>
      </section>

      <p className="text-center font-medium text-lg">
        Join us, draw your first line, and see where the boxes take you.
      </p>
    </div>
  );
}
