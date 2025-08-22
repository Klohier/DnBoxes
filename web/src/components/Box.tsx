import Edge from "./Edge";

interface BoxData {
  box_id: number;
  row: number;
  col: number;
  top_edge: boolean;
  left_edge: boolean;
  right_edge: boolean;
  bottom_edge: boolean;
  completed: boolean | null;
  completed_by: number | null;
}

interface BoxProps {
  box: BoxData;
  userColors: Record<string, string>;
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
}

const Box: React.FC<BoxProps> = ({
  box,
  userColors,
  onEdgeClick,
  currentUserId,
  gameID,
  boxSize,
}) => {
  const {
    box_id,
    row,
    col,
    top_edge,
    left_edge,
    right_edge,
    bottom_edge,
    completed,
    completed_by,
  } = box;

  const x = col * boxSize;
  const y = row * boxSize;

  const handleEdgeClick = (
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge"
  ) => {
    onEdgeClick(gameID, currentUserId, row, col, edge);
  };

  const color = completed_by ? userColors[completed_by] || "gray" : "gray";

  return (
    <g key={box_id}>
      {completed && (
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
            {completed_by}
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
