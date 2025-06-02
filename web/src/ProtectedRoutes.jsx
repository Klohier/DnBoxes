import { Navigate } from "react-router-dom";
import { useUser } from "./UserContext";
import { useAuth } from "./AuthContext";

// // eslint-disable-next-line react/prop-types
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    console.log(" Not Authorized");
    return <Navigate to="/" />;
  }
  console.log("Is Authorized");
  return children;
};

export default ProtectedRoute;
