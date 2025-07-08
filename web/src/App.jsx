import "./App.css";
// import ChatBox from "./components/ChatBox"; // Path to the SvgGrid component
// import Grid from "./components/SvgGrid";
// import { NavLink } from "react-router";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import { UserProvider } from "./UserContext"; // Import the UserProvider
import { AuthProvider } from "./AuthContext";
import { WebSocketProvider } from "./WebSocketContext";
import Game from "./pages/Game";
import ProtectedRoute from "./ProtectedRoutes";
import Home from "./pages/Home";
import Register from "./pages/Registration";
import Login from "./pages/Login";

const ProtectedLayout = () => (
  <ProtectedRoute>
    <WebSocketProvider>
      <Outlet />
    </WebSocketProvider>
  </ProtectedRoute>
);

function App() {
  return (
    <div>
      <BrowserRouter>
        <AuthProvider>
          <UserProvider>
            <Routes>
              <Route path="/" element={<Login />} />
              <Route path="/register" element={<Register />} />

              <Route element={<ProtectedLayout />}>
                <Route path="/home" element={<Home />} />
                <Route path="/game/:gameID" element={<Game />} />
              </Route>
            </Routes>
          </UserProvider>
        </AuthProvider>
      </BrowserRouter>
    </div>
  );
}

export default App;
