import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import Cookies from "js-cookie";
import { useUser } from "../UserContext"; // Import the useUser hook

const Register = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { loginUser } = useUser(); // Get the loginUser function from context

  useEffect(() => {
    const token = Cookies.get("DnB-Session"); // Retrieve session token from cookies
    if (token) {
      // If a session token exists, redirect to home page
      navigate("/home");
    }
  }, [navigate]);

  const handleRegister = async (e) => {
    e.preventDefault();
    setError(""); // Clear previous errors

    try {
      // Create form data in x-www-form-urlencoded format
      const formData = new URLSearchParams();
      formData.append("username", username);
      formData.append("password", password);

      // Make a POST request to your API
      const response = await axios.post(
        "http://localhost:8484/api/v1/users", // Adjust this URL to your registration endpoint
        formData,
        {
          headers: {
            "Content-Type": "application/x-www-form-urlencoded", // Specify the correct content type
          },
          withCredentials: true, // Include cookies if needed
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
