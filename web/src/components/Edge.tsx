import { useState } from "react";

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
  const [hovered, setHovered] = useState(false);
  const strokeColor = active || hovered ? userColor : "gray";
  const strokeDasharray = active || hovered ? "0" : "4,4";
  const cursorStyle = active ? "default" : "pointer";
  return (
    <>
      <line
        x1={x1}
        y1={y1}
        x2={x2}
        y2={y2}
        stroke="transparent"
        strokeWidth="15"
        style={{ cursor: cursorStyle, pointerEvents: "stroke" }}
        onClick={active ? undefined : onClick}
        onMouseEnter={() => {
          setHovered(true);
        }}
        onMouseLeave={() => {
          setHovered(false);
        }}
      />

      <line
        x1={x1}
        y1={y1}
        x2={x2}
        y2={y2}
        stroke={strokeColor}
        strokeDasharray={strokeDasharray}
        strokeWidth="5"
        style={{ pointerEvents: "none" }}
      />
    </>
  );
};

export default Edge;
