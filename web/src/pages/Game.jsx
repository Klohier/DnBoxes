import { useParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Grid from "../components/SvgGrid";
import Chatbox from "../components/ChatBox";
import { useWebSocket } from "../WebSocketContext";

const Game = () => {
  const { gameID } = useParams();
  // const ws = useWebSocket();
  // const navigate = useNavigate();

  // useEffect(() => {
  //   if (!ws) return;

  //   ws.onopen = () => {
  //     console.log("WebSocket connected");
  //     ws.send(JSON.stringify({ type: "player:get" }));
  //   };

  //   ws.onclose = () => {
  //     console.log("WebSocket disconnected");
  //   };

  //   ws.onerror = (err) => {
  //     console.error("WebSocket error:", err);
  //   };

  //   return () => {
  //     ws.removeEventListener("message", ws.onmessage);
  //   };
  // }, []);

  return (
    <div>
      <p>Game ID: {gameID}</p>
      <Grid gameID={gameID} />
      <Chatbox gameID={gameID}></Chatbox>
    </div>
  );
};

export default Game;
