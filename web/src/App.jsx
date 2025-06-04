import "./App.css";
// import ChatBox from "./components/ChatBox"; // Path to the SvgGrid component
import Grid from "./components/SvgGrid";
// import { NavLink } from "react-router";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { UserProvider } from "./UserContext"; // Import the UserProvider
import { AuthProvider } from "./AuthContext";
import { WebSocketProvider } from "./WebSocketContext";
import Game from "./pages/Game";
import ProtectedRoute from "./ProtectedRoutes";
import Home from "./pages/Home";
import Register from "./pages/Registration";
import Login from "./pages/Login";

function App() {
  return (
    <div>
      <BrowserRouter>
        <AuthProvider>
          <UserProvider>
            <WebSocketProvider>
              <Routes>
                <Route path="/" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route
                  path="/home"
                  element={
                    <ProtectedRoute>
                      <Home />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/Grid"
                  element={
                    <ProtectedRoute>
                      <Grid />
                    </ProtectedRoute>
                  }
                />

                <Route
                  path="/game/:gameID"
                  element={
                    <ProtectedRoute>
                      <Game />
                    </ProtectedRoute>
                  }
                />
              </Routes>
            </WebSocketProvider>
          </UserProvider>
        </AuthProvider>
      </BrowserRouter>
    </div>
  );
}

export default App;
