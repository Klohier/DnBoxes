// import { useEffect, useState } from "react";
// import { useUser } from "../UserContext";
// import { useWebSocket } from "../WebSocketContext";
// import { useNavigate } from "react-router-dom";
// import { Toaster, toast } from "sonner";
import Box from "./Box";
import { Box as BoxType } from "@/types/websocket";

interface GridProps {
  gameID: number;
  boxes: BoxType[];
  userColors: Record<number, string>;
  boardSize: number;
  userID: number;
  handleClick: (
    gameID: number,
    userID: number,
    row: number,
    col: number,
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge"
  ) => void;
}

const Grid = ({
  gameID,
  boxes = [],
  userColors,
  boardSize,
  userID,
  handleClick,
}: GridProps) => {
  const boxSize = 70;
  return (
    <div>
      <div className="w-full aspect-square border rounded-lg">
        <svg
          className="w-full h-full"
          viewBox={`-5 -5 ${String(boxSize * boardSize + 10)} ${String(
            boxSize * boardSize + 10
          )}`}
          preserveAspectRatio="xMidYMid meet"
        >
          {boxes.map((box) => (
            <Box
              key={box.box_id}
              box={box}
              userColors={userColors}
              onEdgeClick={handleClick}
              currentUserId={userID}
              gameID={gameID}
              boxSize={boxSize}
              // boardSize={boardSize}
            />
          ))}
        </svg>
      </div>
    </div>
  );
};

export default Grid;
