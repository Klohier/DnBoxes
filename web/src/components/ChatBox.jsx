import { useState, useEffect } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";

// eslint-disable-next-line react/prop-types
const Chatbox = ({ gameID }) => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const { user } = useUser();
  const ws = useWebSocket();

  useEffect(() => {
    const fetchMessages = async () => {
      try {
        // Make an API request to fetch past messages for the given gameID

        const endpoint = gameID
          ? `http://localhost:8484/api/v1/games/${gameID}/chat`
          : `http://localhost:8484/api/v1/chat`;

        const response = await axios.get(endpoint);

        console.log(response);
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
    if (!ws) {
      console.log("No WebSocket connection available");
      return;
    }

    const handleMessage = (event) => {
      const message = JSON.parse(event.data);

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
    };

    // Attach WebSocket event listeners
    ws.addEventListener("message", handleMessage);
    console.log("WebSocket message listener attached");

    return () => {
      // Cleanup event listeners on component unmount
      ws.removeEventListener("message", handleMessage);
      console.log("WebSocket message listener removed");
    };
  }, [gameID, ws]); // Re-run when WebSocket instance changes

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

    if (ws && ws.readyState === WebSocket.OPEN) {
      console.log("Sending message:", message);
      ws.send(JSON.stringify(message));
    } else {
      console.log("WebSocket is not open. ReadyState:", ws.readyState);
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
