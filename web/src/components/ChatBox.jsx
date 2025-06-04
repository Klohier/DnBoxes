import { useState, useEffect } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";

// eslint-disable-next-line react/prop-types
const Chatbox = ({ gameID }) => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const { user } = useUser();
  const { socket, subscribe } = useWebSocket();
  useEffect(() => {
    const fetchMessages = async () => {
      try {
        // Make an API request to fetch past messages for the given gameID

        const endpoint = gameID
          ? `http://localhost:8484/api/v1/games/${gameID}/chat`
          : `http://localhost:8484/api/v1/chat`;

        const response = await axios.get(endpoint);

        // console.log(response);
        if (response.data) {
          setMessages(response.data);
        }
      } catch (error) {
        console.error("Error fetching past messages:", error);
      }
    };

    fetchMessages();
  }, [gameID]);

  useEffect(() => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("No WebSocket connection available");
      return;
    }

    const unsubscribe = subscribe((message) => {
      if (
        message.payload &&
        message.payload.username &&
        message.payload.message &&
        (message.payload.gameID === gameID ||
          (message.payload.gameID === null && gameID === null))
      ) {
        setMessages((prev) => [...prev, message.payload]);
      }

      if (message.type === "chat:new") {
        setMessages((prev) => [...prev, message.payload]);
      }
    });

    console.log("WebSocket message subscription attached");

    return () => {
      unsubscribe();
      console.log("WebSocket message subscription removed");
    };
  }, [gameID, socket, subscribe]);

  const handleSendMessage = () => {
    if (newMessage.trim() === "") return;

    const message = {
      type: "chat:new",
      payload: {
        userID: parseInt(user.userID),
        username: user.username,
        gameID: parseInt(gameID),
        message: newMessage,
        timestamp: new Date().toISOString(),
      },
    };

    if (socket && socket.readyState === WebSocket.OPEN) {
      console.log("Sending message:", message);
      socket.send(JSON.stringify(message));
    } else if (socket) {
      console.log("WebSocket is not open. ReadyState:", socket.readyState);
    } else {
      console.log("WebSocket is not available");
    }

    setNewMessage("");
  };

  return (
    <div style={{ border: "1px solid #ccc", padding: "10px", width: "300px" }}>
      <div
        style={{
          maxHeight: "200px",
          overflowY: "scroll",
          marginBottom: "10px",
        }}
      >
        {messages.map((msg, index) => (
          <div key={index}>
            <strong>{msg.username}:</strong> {msg.message}
          </div>
        ))}
      </div>
      <input
        type="text"
        value={newMessage}
        onChange={(e) => setNewMessage(e.target.value)}
        placeholder="Type a message"
        style={{ width: "100%", marginBottom: "5px" }}
      />
      <button onClick={handleSendMessage}>Send</button>
    </div>
  );
};

export default Chatbox;
