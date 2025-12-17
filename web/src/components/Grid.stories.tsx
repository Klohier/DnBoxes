import type { Meta, StoryObj } from "@storybook/react-vite";

import Grid from "./Grid";
import { Box } from "@/types/websocket";

const meta = {
  component: Grid,
} satisfies Meta<typeof Grid>;

export default meta;

const generateMockBoxes = (boardSize: number): Box[] => {
  const boxes: Box[] = [];

  for (let row = 0; row < boardSize; row++) {
    for (let col = 0; col < boardSize; col++) {
      boxes.push({
        row,
        col,
        top_edge: false,
        left_edge: false,
        right_edge: false,
        bottom_edge: false,
        owner_turn: null, // Changed from completed_by
      });
    }
  }

  return boxes;
};

// Mock mapping from turn_order to user_id
const mockTurnToUserIdMap: Record<number, number> = {
  0: 1, // Player in turn position 0 is user_id 1
  1: 2, // Player in turn position 1 is user_id 2
  2: 3, // Player in turn position 2 is user_id 3
  3: 4, // Player in turn position 3 is user_id 4
};

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    gameID: 123,
    boardSize: 5,
    boxes: generateMockBoxes(5),
    userColors: {
      1: "#ff0000", // red
      2: "#0000ff", // blue
      3: "#00ff00", // green
      4: "#ff00ff", // purple
    },
    userID: 1,
    turnToUserIdMap: mockTurnToUserIdMap,
    handleClick: (gameID, userID, row, col, edge) => {
      console.log(
        "Clicked edge",
        edge,
        "at position",
        { row, col },
        "by user",
        userID
      );
    },
  },
  argTypes: {
    boardSize: {
      control: { type: "number", min: 1, max: 10 },
    },
  },
  render: (args) => {
    return (
      <Grid
        {...args}
        boxes={generateMockBoxes(args.boardSize)}
        boardSize={args.boardSize}
      />
    );
  },
};

export const WithCompletedBoxes: Story = {
  args: {
    gameID: 123,
    boardSize: 5,
    boxes: generateMockBoxes(5).map((box, index) => {
      // Complete first 3 boxes
      if (index < 3) {
        return {
          ...box,
          top_edge: true,
          right_edge: true,
          bottom_edge: true,
          left_edge: true,
          owner_turn: index % 2, // Alternate between turn 0 and 1
        };
      }
      return box;
    }),
    userColors: {
      1: "#ff0000",
      2: "#0000ff",
    },
    userID: 1,
    turnToUserIdMap: mockTurnToUserIdMap,
    handleClick: (gameID, userID, row, col, edge) => {
      console.log("Clicked edge", edge, "at position", { row, col });
    },
  },
};

export const PartiallyFilledGrid: Story = {
  args: {
    gameID: 123,
    boardSize: 5,
    boxes: generateMockBoxes(5).map((box) => {
      // Randomly fill some edges
      return {
        ...box,
        top_edge: Math.random() > 0.5,
        right_edge: Math.random() > 0.5,
        bottom_edge: Math.random() > 0.5,
        left_edge: Math.random() > 0.5,
        owner_turn: Math.random() > 0.7 ? Math.floor(Math.random() * 2) : null,
      };
    }),
    userColors: {
      1: "#ff0000",
      2: "#0000ff",
    },
    userID: 1,
    turnToUserIdMap: mockTurnToUserIdMap,
    handleClick: (gameID, userID, row, col, edge) => {
      console.log("Clicked edge", edge, "at position", { row, col });
    },
  },
};

export const SmallGrid: Story = {
  args: {
    gameID: 123,
    boardSize: 3,
    boxes: generateMockBoxes(3),
    userColors: {
      1: "#ff0000",
      2: "#0000ff",
    },
    userID: 1,
    turnToUserIdMap: mockTurnToUserIdMap,
    handleClick: (gameID, userID, row, col, edge) => {
      console.log("Clicked edge", edge, "at position", { row, col });
    },
  },
};

export const LargeGrid: Story = {
  args: {
    gameID: 123,
    boardSize: 8,
    boxes: generateMockBoxes(8),
    userColors: {
      1: "#ff0000",
      2: "#0000ff",
      3: "#00ff00",
      4: "#ff00ff",
    },
    userID: 1,
    turnToUserIdMap: mockTurnToUserIdMap,
    handleClick: (gameID, userID, row, col, edge) => {
      console.log("Clicked edge", edge, "at position", { row, col });
    },
  },
};
