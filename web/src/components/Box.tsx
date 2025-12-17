import Edge from "./Edge";

interface BoxData {
  row: number;
  col: number;
  top_edge: boolean;
  left_edge: boolean;
  right_edge: boolean;
  bottom_edge: boolean;
  owner_turn: number | null; // Changed from completed_by
}

interface BoxProps {
  box: BoxData;
  userColors: Record<number, string>; // Changed key type from string to number
  onEdgeClick: (
    gameID: number,
    userID: number,
    row: number,
    col: number,
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge"
  ) => void;
  currentUserId: number;
  gameID: number;
  boxSize: number;
  turnToUserIdMap: Record<number, number>; // Added to map turn_order to user_id
}

const Box: React.FC<BoxProps> = ({
  box,
  userColors,
  onEdgeClick,
  currentUserId,
  gameID,
  boxSize,
  turnToUserIdMap,
}) => {
  const { row, col, top_edge, left_edge, right_edge, bottom_edge, owner_turn } =
    box;

  const x = col * boxSize;
  const y = row * boxSize;

  const handleEdgeClick = (
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge"
  ) => {
    onEdgeClick(gameID, currentUserId, row, col, edge);
  };

  // Check if box is completed
  const isCompleted = owner_turn !== null;

  // Get user_id from turn_order, then get color
  const ownerId = owner_turn !== null ? turnToUserIdMap[owner_turn] : null;
  const color = ownerId ? userColors[ownerId] || "gray" : "gray";

  return (
    <g key={`${row}-${col}`}>
      {isCompleted && (
        <>
          <rect x={x} y={y} width={boxSize} height={boxSize} fill={color} />

          <text
            x={x + boxSize / 2}
            y={y + boxSize / 2}
            textAnchor="middle"
            alignmentBaseline="middle"
            fontSize="12"
            fill="white"
          >
            {ownerId}
          </text>
        </>
      )}

      <Edge
        x1={x}
        y1={y}
        x2={x + boxSize}
        y2={y}
        active={top_edge}
        onClick={() => {
          handleEdgeClick("top_edge");
        }}
        userColor={color}
      />
      <Edge
        x1={x}
        y1={y}
        x2={x}
        y2={y + boxSize}
        active={left_edge}
        onClick={() => {
          handleEdgeClick("left_edge");
        }}
        userColor={color}
      />

      <Edge
        x1={x + boxSize}
        y1={y}
        x2={x + boxSize}
        y2={y + boxSize}
        active={right_edge}
        onClick={() => {
          handleEdgeClick("right_edge");
        }}
        userColor={color}
      />
      <Edge
        x1={x}
        y1={y + boxSize}
        x2={x + boxSize}
        y2={y + boxSize}
        active={bottom_edge}
        onClick={() => {
          handleEdgeClick("bottom_edge");
        }}
        userColor={color}
      />
    </g>
  );
};

export default Box;
