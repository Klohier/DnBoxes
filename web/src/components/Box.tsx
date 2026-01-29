import Edge from "./Edge";

interface BoxData {
  row: number;
  col: number;
  top_edge: boolean;
  left_edge: boolean;
  right_edge: boolean;
  bottom_edge: boolean;
  owner_turn: number | null;
}

interface BoxProps {
  box: BoxData;
  userColors: Record<number, string>;
  onEdgeClick: (
    gameID: number,
    userID: number,
    row: number,
    col: number,
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge",
  ) => void;
  currentUserId: number;
  gameID: number;
  boxSize: number;
  turnToUserIdMap: Record<number, number>;
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
  const centerX = x + boxSize / 2;
  const centerY = y + boxSize / 2;

  const handleEdgeClick = (
    edge: "top_edge" | "left_edge" | "right_edge" | "bottom_edge",
  ) => {
    onEdgeClick(gameID, currentUserId, row, col, edge);
  };

  const isCompleted = owner_turn !== null;
  const ownerId = owner_turn !== null ? turnToUserIdMap[owner_turn] : null;
  const color = ownerId ? userColors[ownerId] || "gray" : "gray";

  return (
    <g key={`${row}-${col}`}>
      {isCompleted && (
        <>
          {/* Main filled box - animates once on mount */}
          <rect
            x={x + 2}
            y={y + 2}
            width={boxSize - 4}
            height={boxSize - 4}
            fill={color}
            fillOpacity={0.85}
            rx="4"
            style={{
              transformOrigin: `${centerX}px ${centerY}px`,
              animation: "boxPop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)",
            }}
          />

          {/* Inner bright ring - animates once then disappears */}
          <circle
            cx={centerX}
            cy={centerY}
            r={boxSize * 0.2}
            fill="none"
            stroke="#ffffff"
            strokeWidth="4"
            style={{
              animation: "expandRingInner 0.6s ease-out forwards",
              pointerEvents: "none",
            }}
          />

          {/* Outer colored ring - animates once then disappears */}
          <circle
            cx={centerX}
            cy={centerY}
            r={boxSize * 0.25}
            fill="none"
            stroke={color}
            strokeWidth="5"
            style={{
              animation: "expandRingOuter 0.8s ease-out forwards",
              filter: "brightness(1.3)",
              pointerEvents: "none",
            }}
          />

          {/* Score text - animates once then disappears */}
          <text
            x={centerX}
            y={centerY}
            textAnchor="middle"
            dominantBaseline="middle"
            fontSize="32"
            fontWeight="bold"
            fill="#ffffff"
            style={{
              filter: "drop-shadow(0 2px 4px rgba(0,0,0,0.8))",
              animation: "scoreAppear 0.8s ease-out forwards",
              pointerEvents: "none",
            }}
          >
            +1
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
        userColor={color}
      />
      <Edge
        x1={x}
        y1={y}
        x2={x}
        y2={y + boxSize}
        active={left_edge}
        onClick={() => handleEdgeClick("left_edge")}
        userColor={color}
      />
      <Edge
        x1={x + boxSize}
        y1={y}
        x2={x + boxSize}
        y2={y + boxSize}
        active={right_edge}
        onClick={() => handleEdgeClick("right_edge")}
        userColor={color}
      />
      <Edge
        x1={x}
        y1={y + boxSize}
        x2={x + boxSize}
        y2={y + boxSize}
        active={bottom_edge}
        onClick={() => handleEdgeClick("bottom_edge")}
        userColor={color}
      />

      <style>{`
        @keyframes boxPop {
          0% {
            transform: scale(0.5);
            opacity: 0;
          }
          50% {
            transform: scale(1.1);
          }
          100% {
            transform: scale(1);
            opacity: 1;
          }
        }
        
        @keyframes expandRingInner {
          0% {
            r: ${boxSize * 0.2};
            stroke-opacity: 1;
            stroke-width: 4;
          }
          100% {
            r: ${boxSize * 0.55};
            stroke-opacity: 0;
            stroke-width: 2;
          }
        }
        
        @keyframes expandRingOuter {
          0% {
            r: ${boxSize * 0.25};
            stroke-opacity: 1;
            stroke-width: 5;
          }
          100% {
            r: ${boxSize * 0.7};
            stroke-opacity: 0;
            stroke-width: 1;
          }
        }
        
        @keyframes scoreAppear {
          0% {
            opacity: 0;
            transform: scale(0.5);
          }
          40% {
            opacity: 1;
            transform: scale(1.15);
          }
          100% {
            opacity: 0;
            transform: scale(1);
          }
        }
      `}</style>
    </g>
  );
};

export default Box;
