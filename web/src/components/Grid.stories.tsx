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
      const box_id = row * boardSize + col;
      boxes.push({
        box_id,
        game_id: 1,
        top_edge: false,
        left_edge: false,
        right_edge: false,
        bottom_edge: false,
        row,
        col,
        completed: null,
        completed_by: null,
      });
    }
  }

  return boxes;
};

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    gameID: 123,
    boardSize: 1,
    boxes: generateMockBoxes(5),
    userColors: {
      1: "#ff0000",
      2: "#0000ff",
    },
    userID: 1,
    handleClick: (edge, boxID) => {
      console.log("Clicked edge", edge, "on box", boxID);
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
