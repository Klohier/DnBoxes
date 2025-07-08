/* eslint-disable react/prop-types */
import Edge from "./Edge";

const Box = ({
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

  const handleEdgeClick = (edge) => {
    onEdgeClick(gameID, currentUserId, row, col, edge);
  };

  return (
    <g key={box_id}>
      {completed && (
        <>
          <rect
            x={x}
            y={y}
            width={boxSize}
            height={boxSize}
            fill={userColors[completed_by] || "gray"}
          />

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
        onClick={() => handleEdgeClick("top_edge")}
        userColor={userColors[completed_by] || "gray"}
      />
      <Edge
        x1={x}
        y1={y}
        x2={x}
        y2={y + boxSize}
        active={left_edge}
        onClick={() => handleEdgeClick("left_edge")}
        userColor={userColors[completed_by] || "gray"}
      />

      <Edge
        x1={x + boxSize}
        y1={y}
        x2={x + boxSize}
        y2={y + boxSize}
        active={right_edge}
        onClick={() => handleEdgeClick("right_edge")}
        userColor={userColors[completed_by] || "gray"}
      />
      <Edge
        x1={x}
        y1={y + boxSize}
        x2={x + boxSize}
        y2={y + boxSize}
        active={bottom_edge}
        onClick={() => handleEdgeClick("bottom_edge")}
        userColor={userColors[completed_by] || "gray"}
      />
    </g>
  );
};

export default Box;
