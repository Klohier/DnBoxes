import { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { useUser } from "../UserContext";

const Register = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { loginUser } = useUser();
  const apiUrl = import.meta.env.VITE_API_URL || "localhost:8484";

  const handleRegister = async (e) => {
    e.preventDefault();
    setError("");
    try {
      const formData = new URLSearchParams();
      formData.append("username", username);
      formData.append("password", password);

      const response = await axios.post(
        `http://${apiUrl}/api/v1/users`,
        formData,
        {
          headers: {
            "Content-Type": "application/x-www-form-urlencoded",
          },
          withCredentials: true,
        }
      );

      if (response.status === 201) {
        // Redirect to the home page and log the user in automatically
        loginUser(response.data);
        setError("Account Created, Click Login to Access");
      } else {
        setError("Registration failed. Please try again.");
      }
    } catch (err) {
      console.error("Registration error:", err);
      setError("Failed to register. Please try again.");
    }
  };
  const handleLogin = () => {
    navigate("/");
  };

  return (
    <div>
      <h2>Register</h2>
      <form onSubmit={handleRegister}>
        <div>
          <label>Username:</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Password:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>

        {error && <div style={{ color: "red" }}>{error}</div>}
        <button type="submit">Register</button>
        <button onClick={handleLogin}>Login</button>
      </form>
    </div>
  );
};

export default Register;
