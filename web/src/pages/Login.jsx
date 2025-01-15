import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../AuthContext";
const Login = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { login, token } = useAuth();

  useEffect(() => {
    if (token) {
      // If a session token exists, redirect to home page
      navigate("/home");
    }
  }, [navigate, token]);

  const handleLogin = async (e) => {
    e.preventDefault();
    setError("");

    try {
      const formData = new URLSearchParams();
      formData.append("username", username);
      formData.append("password", password);

      await login(formData);
      //   const response = await axios.post(
      //     "http://localhost:8484/api/v1/login",
      //     formData,
      //     {
      //       headers: {
      //         "Content-Type": "application/x-www-form-urlencoded",
      //       },
      //       withCredentials: true,
      //     }
      //   );

      //   if (response.status === 200) {
      //     loginUser(response.data);
      //     navigate("/home");
      //   } else {
      //     setError("Invalid username or password");
      //   }
    } catch (err) {
      console.log("Login error:", err);
      setError("Failed to log in. Please try again.");
    }
  };

  const handleRegister = () => {
    navigate("/register");
  };
  return (
    <div>
      <h2>Login</h2>
      <form onSubmit={handleLogin}>
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
        <button type="submit">Login</button>
      </form>
      <button onClick={handleRegister}>Sign Up</button>
    </div>
  );
};

export default Login;
