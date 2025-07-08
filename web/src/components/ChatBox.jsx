import { useState, useEffect } from "react";
import { useUser } from "../UserContext";
import { useWebSocket } from "../WebSocketContext";
import axios from "axios";

// eslint-disable-next-line react/prop-types
const Chatbox = ({ sessionID }) => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const { user } = useUser();
  const { socket, subscribe } = useWebSocket();
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";

  useEffect(() => {
    if (!sessionID || !socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    const fetchMessages = async () => {
      try {
        // Make an API request to fetch past messages for the given gameID

        const endpoint = `http://${apiUrl}/api/v1/chat?sessionID=${sessionID}`;

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
  }, [sessionID, socket]);

  useEffect(() => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log("No WebSocket connection available");
      return;
    }

    const unsubscribe = subscribe((message) => {
      // if (
      //   message.payload &&
      //   message.payload.username &&
      //   message.payload.message &&
      //   (message.payload.session_id === sessionID ||
      //     (message.payload.session_id === null && sessionID === null))
      // ) {
      //   setMessages((prev) => [...prev, message.payload]);
      // }

      if (message.type === "chat:new") {
        setMessages((prev) => [...prev, message.payload]);
      }
    });

    console.log("WebSocket message subscription attached");

    return () => {
      unsubscribe();
      console.log("WebSocket message subscription removed");
    };
  }, [sessionID, socket, subscribe]);

  const handleSendMessage = () => {
    if (newMessage.trim() === "") return;

    const message = {
      type: "chat:new",
      payload: {
        userID: parseInt(user.userID),
        username: user.username,
        session_id: parseInt(sessionID),
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
