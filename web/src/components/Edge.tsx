interface EdgeProps {
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  active: boolean;
  onClick: () => void;
  userColor: string;
}

const Edge: React.FC<EdgeProps> = ({
  x1,
  y1,
  x2,
  y2,
  active,
  onClick,
  userColor,
}) => {
  const strokeColor = active ? "#e5e7eb" : "#4b5563";
  const strokeDasharray = active ? "0" : "4,4";
  const cursorStyle = active ? "default" : "pointer";

  // Calculate line length for drawing animation
  const lineLength = Math.sqrt(Math.pow(x2 - x1, 2) + Math.pow(y2 - y1, 2));

  return (
    <>
      {/* Invisible hitbox for hover/click */}
      <line
        x1={x1}
        y1={y1}
        x2={x2}
        y2={y2}
        stroke="transparent"
        strokeWidth="15"
        style={{ cursor: cursorStyle, pointerEvents: "stroke" }}
        onClick={active ? undefined : onClick}
        className="edge-hover"
      />

      {/* Visible line with drawing animation when active */}
      {active ? (
        <>
          {/* Drawing line animation */}
          <line
            x1={x1}
            y1={y1}
            x2={x2}
            y2={y2}
            stroke={strokeColor}
            strokeWidth="5"
            strokeLinecap="round"
            strokeDasharray={lineLength}
            strokeDashoffset={lineLength}
            style={{
              pointerEvents: "none",
              animation: `drawLine 0.4s ease-out forwards`,
            }}
          />

          {/* Glow effect 1 */}
          <line
            x1={x1}
            y1={y1}
            x2={x2}
            y2={y2}
            stroke="#fbbf24"
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={lineLength}
            strokeDashoffset={lineLength}
            style={{
              pointerEvents: "none",
              animation: `drawLineGlow 0.5s ease-out forwards`,
            }}
          />

          {/* Glow effect 2 */}
          <line
            x1={x1}
            y1={y1}
            x2={x2}
            y2={y2}
            stroke="#f59e0b"
            strokeWidth="15"
            strokeLinecap="round"
            strokeDasharray={lineLength}
            strokeDashoffset={lineLength}
            style={{
              pointerEvents: "none",
              animation: `drawLineGlowOuter 0.6s ease-out forwards`,
            }}
          />
        </>
      ) : (
        <line
          x1={x1}
          y1={y1}
          x2={x2}
          y2={y2}
          stroke={strokeColor}
          strokeDasharray={strokeDasharray}
          strokeWidth="5"
          strokeLinecap="round"
          className="edge-line"
          style={{
            pointerEvents: "none",
            transition: "stroke 0.2s ease, stroke-dasharray 0.2s ease",
          }}
        />
      )}

      <style>
        {`
          @keyframes drawLine {
            from {
              stroke-dashoffset: ${lineLength};
            }
            to {
              stroke-dashoffset: 0;
            }
          }
          
          @keyframes drawLineGlow {
            0% {
              stroke-dashoffset: ${lineLength};
              opacity: 0.8;
            }
            100% {
              stroke-dashoffset: 0;
              opacity: 0;
            }
          }
          
          @keyframes drawLineGlowOuter {
            0% {
              stroke-dashoffset: ${lineLength};
              opacity: 0.6;
            }
            100% {
              stroke-dashoffset: 0;
              opacity: 0;
            }
          }

          .edge-hover:hover + .edge-line {
            stroke: ${userColor};
            stroke-dasharray: 0;
          }
        `}
      </style>
    </>
  );
};

export default Edge;
