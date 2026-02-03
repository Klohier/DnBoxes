import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { fetchGame } from "@/api/fetchGame";
import {
  fetchGameEvents,
  DomainEvent,
  GameCreatedPayload,
  MoveAppliedPayload,
  BoxCompletedPayload,
  TurnPassedPayload,
} from "@/api/fetchGames";
import { useState, useEffect, useCallback, useMemo, useRef } from "react";
import Grid from "@/components/Grid";
import { Box } from "@/types/websocket";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  ChevronsLeft,
  ChevronsRight,
  ArrowLeft,
  Users,
  Trophy,
} from "lucide-react";

export const Route = createFileRoute("/_authenticated/replay/$gameID")({
  component: ReplayPage,
});

type EdgeField = "top_edge" | "right_edge" | "bottom_edge" | "left_edge";

function edgeNameToField(edge: string): EdgeField {
  switch (edge) {
    case "top":
      return "top_edge";
    case "right":
      return "right_edge";
    case "bottom":
      return "bottom_edge";
    case "left":
      return "left_edge";
    default:
      return "top_edge";
  }
}

function makeEmptyGrid(size: number): Box[][] {
  const grid: Box[][] = [];
  for (let r = 0; r < size; r++) {
    grid[r] = [];
    for (let c = 0; c < size; c++) {
      grid[r][c] = {
        row: r,
        col: c,
        top_edge: false,
        right_edge: false,
        bottom_edge: false,
        left_edge: false,
        owner_turn: null,
      };
    }
  }
  return grid;
}

function setEdge(grid: Box[][], boardSize: number, row: number, col: number, edge: EdgeField) {
  const box = grid[row][col];
  switch (edge) {
    case "top_edge":
      box.top_edge = true;
      if (row > 0) grid[row - 1][col].bottom_edge = true;
      break;
    case "right_edge":
      box.right_edge = true;
      if (col < boardSize - 1) grid[row][col + 1].left_edge = true;
      break;
    case "bottom_edge":
      box.bottom_edge = true;
      if (row < boardSize - 1) grid[row + 1][col].top_edge = true;
      break;
    case "left_edge":
      box.left_edge = true;
      if (col > 0) grid[row][col - 1].right_edge = true;
      break;
  }
}

interface ReplayState {
  grid: Box[][];
  scores: Record<number, number>;
  currentTurn: number;
  lastMove?: MoveAppliedPayload;
}

function buildReplayStates(domainEvents: DomainEvent[]): ReplayState[] {
  const states: ReplayState[] = [];

  let boardSize = 0;
  let grid: Box[][] = [];
  let scores: Record<number, number> = {};
  let currentTurn = 0;
  let hasPendingMove = false;
  let lastMove: MoveAppliedPayload | undefined;

  for (const event of domainEvents) {
    // If a new MoveApplied arrives while there's a pending move snapshot
    // (previous move completed boxes without a TurnPassed), push that snapshot now
    if (event.type === "game.move_applied" && hasPendingMove) {
      states.push({
        grid: JSON.parse(JSON.stringify(grid)),
        scores: { ...scores },
        currentTurn,
        lastMove,
      });
      hasPendingMove = false;
    }

    switch (event.type) {
      case "game.created": {
        const p = event.payload as GameCreatedPayload;
        boardSize = p.board_size;
        grid = makeEmptyGrid(boardSize);
        scores = {};
        currentTurn = 0;
        states.push({
          grid: JSON.parse(JSON.stringify(grid)),
          scores: { ...scores },
          currentTurn,
        });
        break;
      }
      case "game.move_applied": {
        const p = event.payload as MoveAppliedPayload;
        const edgeField = edgeNameToField(p.edge);
        setEdge(grid, boardSize, p.row, p.col, edgeField);
        lastMove = p;
        hasPendingMove = true;
        break;
      }
      case "game.box_completed": {
        const p = event.payload as BoxCompletedPayload;
        grid[p.row][p.col].owner_turn = p.owner_turn;
        scores[p.owner_turn] = (scores[p.owner_turn] || 0) + 1;
        break;
      }
      case "game.turn_passed": {
        const p = event.payload as TurnPassedPayload;
        currentTurn = p.next_turn;
        states.push({
          grid: JSON.parse(JSON.stringify(grid)),
          scores: { ...scores },
          currentTurn,
          lastMove,
        });
        hasPendingMove = false;
        break;
      }
      case "game.ended":
      case "game.forfeited": {
        states.push({
          grid: JSON.parse(JSON.stringify(grid)),
          scores: { ...scores },
          currentTurn,
          lastMove,
        });
        hasPendingMove = false;
        break;
      }
    }
  }

  // Flush any pending move at the end (shouldn't normally happen)
  if (hasPendingMove) {
    states.push({
      grid: JSON.parse(JSON.stringify(grid)),
      scores: { ...scores },
      currentTurn,
      lastMove,
    });
  }

  return states;
}

const SPEED_OPTIONS = [
  { label: "0.5x", value: 2000 },
  { label: "1x", value: 1000 },
  { label: "2x", value: 500 },
  { label: "4x", value: 250 },
];

function ReplayPage() {
  const { gameID } = Route.useParams();
  const navigate = useNavigate();

  const { data: gameState, isLoading: gameLoading } = useQuery({
    queryKey: ["game", gameID],
    queryFn: () => fetchGame(gameID),
  });

  const { data: events, isLoading: eventsLoading } = useQuery({
    queryKey: ["gameEvents", gameID],
    queryFn: () => fetchGameEvents(Number(gameID)),
  });

  const [currentStep, setCurrentStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speedIndex, setSpeedIndex] = useState(1);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const replayStates = useMemo(() => {
    if (!events || events.length === 0) return [];
    return buildReplayStates(events);
  }, [events]);

  const totalSteps = replayStates.length;

  const currentState = replayStates[currentStep] || null;

  const userColors = useMemo(() => {
    if (!gameState) return {};
    const colors = ["red", "blue", "green", "purple", "orange", "pink"];
    const colorMap: Record<number, string> = {};
    gameState.players.forEach((player) => {
      colorMap[player.user_id] = colors[player.turn_order % colors.length];
    });
    return colorMap;
  }, [gameState]);

  const turnToUserIdMap = useMemo(() => {
    if (!gameState) return {};
    return gameState.players.reduce(
      (acc, player) => {
        acc[player.turn_order] = player.user_id;
        return acc;
      },
      {} as Record<number, number>,
    );
  }, [gameState]);

  // Stop playback
  const stopPlayback = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    setIsPlaying(false);
  }, []);

  // Start playback
  const startPlayback = useCallback(() => {
    if (currentStep >= totalSteps - 1) {
      setCurrentStep(0);
    }
    setIsPlaying(true);
  }, [currentStep, totalSteps]);

  // Handle auto-play interval
  useEffect(() => {
    if (!isPlaying) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    intervalRef.current = setInterval(() => {
      setCurrentStep((prev) => {
        if (prev >= totalSteps - 1) {
          stopPlayback();
          return prev;
        }
        return prev + 1;
      });
    }, SPEED_OPTIONS[speedIndex].value);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [isPlaying, speedIndex, totalSteps, stopPlayback]);

  const stepForward = useCallback(() => {
    stopPlayback();
    setCurrentStep((prev) => Math.min(prev + 1, totalSteps - 1));
  }, [totalSteps, stopPlayback]);

  const stepBackward = useCallback(() => {
    stopPlayback();
    setCurrentStep((prev) => Math.max(prev - 1, 0));
  }, [stopPlayback]);

  const goToStart = useCallback(() => {
    stopPlayback();
    setCurrentStep(0);
  }, [stopPlayback]);

  const goToEnd = useCallback(() => {
    stopPlayback();
    setCurrentStep(totalSteps - 1);
  }, [totalSteps, stopPlayback]);

  const togglePlay = useCallback(() => {
    if (isPlaying) {
      stopPlayback();
    } else {
      startPlayback();
    }
  }, [isPlaying, stopPlayback, startPlayback]);

  const cycleSpeed = useCallback(() => {
    setSpeedIndex((prev) => (prev + 1) % SPEED_OPTIONS.length);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case " ":
          e.preventDefault();
          togglePlay();
          break;
        case "ArrowRight":
          e.preventDefault();
          stepForward();
          break;
        case "ArrowLeft":
          e.preventDefault();
          stepBackward();
          break;
        case "Home":
          e.preventDefault();
          goToStart();
          break;
        case "End":
          e.preventDefault();
          goToEnd();
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [togglePlay, stepForward, stepBackward, goToStart, goToEnd]);

  // No-op click handler for the grid (replay mode is view-only)
  const noopClick = useCallback(() => {}, []);

  if (gameLoading || eventsLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <p className="text-lg text-gray-300">Loading replay...</p>
      </div>
    );
  }

  if (!gameState || !events) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <p className="text-lg text-gray-300">Game not found</p>
      </div>
    );
  }

  if (replayStates.length <= 1) {
    return (
      <div className="min-h-screen bg-gray-900 p-4">
        <div className="max-w-3xl mx-auto space-y-6 text-center">
          <h1 className="text-3xl font-bold text-white">
            Replay: Game #{gameID}
          </h1>
          <p className="text-gray-400">
            No event data available for this game. Event recording was added
            after this game was played.
          </p>
          <Button
            variant="outline"
            onClick={() => navigate({ to: "/history" })}
            className="border-gray-600 text-gray-300"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to History
          </Button>
        </div>
      </div>
    );
  }

  const flattenedBoxes = currentState ? currentState.grid.flat() : [];

  // Current move info from domain events
  const lastMove = currentState?.lastMove;
  const activePlayer = lastMove
    ? gameState.players.find((p) => p.turn_order === lastMove.turn_order)
    : null;

  const winnerPlayer = gameState.players.find(
    (p) => p.user_id === gameState.winner_id,
  );

  const isAtEnd = currentStep >= totalSteps - 1;

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate({ to: "/history" })}
              className="text-gray-400 hover:text-white"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              History
            </Button>
            <h1 className="text-xl font-bold text-white">
              Replay: Game #{gameID}
            </h1>
            <Badge variant="outline" className="text-gray-400">
              {gameState.board_size}x{gameState.board_size}
            </Badge>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* Left - Game Board */}
          <div className="lg:col-span-2 space-y-4">
            {/* Move Info Bar */}
            <Card className="bg-gray-800 border-gray-700">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    {currentStep === 0 ? (
                      <p className="text-gray-400">Game start</p>
                    ) : isAtEnd ? (
                      <div className="flex items-center gap-2">
                        <Trophy className="h-5 w-5 text-yellow-500" />
                        <p className="text-white font-semibold">
                          {winnerPlayer
                            ? `${winnerPlayer.username} wins!`
                            : "Draw!"}
                        </p>
                      </div>
                    ) : activePlayer && lastMove ? (
                      <p className="text-white">
                        <span
                          className="font-semibold"
                          style={{
                            color: userColors[activePlayer.user_id] || "white",
                          }}
                        >
                          {activePlayer.username}
                        </span>{" "}
                        placed{" "}
                        <span className="text-gray-300">
                          {lastMove.edge} edge at (
                          {lastMove.row}, {lastMove.col})
                        </span>
                      </p>
                    ) : null}
                  </div>
                  <Badge variant="secondary">
                    Move {currentStep} / {totalSteps - 1}
                  </Badge>
                </div>
              </CardContent>
            </Card>

            {/* Board */}
            <Card className="bg-gray-800 border-gray-700">
              <CardContent className="p-6 flex justify-center">
                <div className="w-full max-w-[650px]">
                  {currentState && (
                    <Grid
                      gameID={gameState.game_id}
                      boxes={flattenedBoxes}
                      userColors={userColors}
                      boardSize={gameState.board_size}
                      userID={0}
                      handleClick={noopClick}
                      turnToUserIdMap={turnToUserIdMap}
                    />
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Playback Controls */}
            <Card className="bg-gray-800 border-gray-700">
              <CardContent className="p-4 space-y-3">
                {/* Progress bar */}
                <div className="relative">
                  <input
                    type="range"
                    min={0}
                    max={totalSteps - 1}
                    value={currentStep}
                    onChange={(e) => {
                      stopPlayback();
                      setCurrentStep(Number(e.target.value));
                    }}
                    className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-500"
                  />
                </div>

                {/* Controls */}
                <div className="flex items-center justify-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={goToStart}
                    disabled={currentStep === 0}
                    className="text-gray-400 hover:text-white"
                  >
                    <ChevronsLeft className="h-5 w-5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={stepBackward}
                    disabled={currentStep === 0}
                    className="text-gray-400 hover:text-white"
                  >
                    <SkipBack className="h-5 w-5" />
                  </Button>
                  <Button
                    variant="default"
                    size="sm"
                    onClick={togglePlay}
                    className="px-6"
                  >
                    {isPlaying ? (
                      <Pause className="h-5 w-5" />
                    ) : (
                      <Play className="h-5 w-5" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={stepForward}
                    disabled={isAtEnd}
                    className="text-gray-400 hover:text-white"
                  >
                    <SkipForward className="h-5 w-5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={goToEnd}
                    disabled={isAtEnd}
                    className="text-gray-400 hover:text-white"
                  >
                    <ChevronsRight className="h-5 w-5" />
                  </Button>
                  <div className="ml-4 border-l border-gray-700 pl-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={cycleSpeed}
                      className="border-gray-600 text-gray-300 hover:text-white min-w-[60px]"
                    >
                      {SPEED_OPTIONS[speedIndex].label}
                    </Button>
                  </div>
                </div>

                <p className="text-center text-xs text-gray-500">
                  Space to play/pause, Arrow keys to step, Home/End to
                  jump
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Right - Players */}
          <div className="lg:col-span-1 space-y-4">
            <Card className="bg-gray-800 border-gray-700">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <Users className="h-5 w-5" />
                  Players
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[...gameState.players]
                    .sort((a, b) => {
                      const scoreA = currentState?.scores[a.turn_order] || 0;
                      const scoreB = currentState?.scores[b.turn_order] || 0;
                      return scoreB - scoreA;
                    })
                    .map((player) => {
                      const score =
                        currentState?.scores[player.turn_order] || 0;
                      const isActive =
                        lastMove?.turn_order === player.turn_order &&
                        currentStep > 0 &&
                        !isAtEnd;
                      return (
                        <div
                          key={player.user_id}
                          className={`flex items-center justify-between p-3 rounded-lg transition-colors ${
                            isActive
                              ? "bg-blue-500/20 border border-blue-500/50"
                              : "bg-gray-700/50"
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            {isAtEnd &&
                              player.user_id === gameState.winner_id && (
                                <Trophy className="h-5 w-5 text-yellow-500" />
                              )}
                            <div
                              className="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold text-sm"
                              style={{
                                backgroundColor:
                                  userColors[player.user_id] || "#666",
                              }}
                            >
                              {player.username.charAt(0).toUpperCase()}
                            </div>
                            <div>
                              <span className="text-white font-medium">
                                {player.username}
                              </span>
                              {isActive && (
                                <Badge
                                  variant="outline"
                                  className="ml-2 text-xs"
                                >
                                  Last move
                                </Badge>
                              )}
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="text-2xl font-bold text-white">
                              {score}
                            </p>
                            <p className="text-xs text-gray-400">points</p>
                          </div>
                        </div>
                      );
                    })}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
