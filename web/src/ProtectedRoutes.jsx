import { Navigate } from "react-router-dom";
import { useUser } from "./UserContext";

// // eslint-disable-next-line react/prop-types
const ProtectedRoute = ({ children }) => {
  const { user, loading } = useUser();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!user) {
    console.log("Not Authorized");
    return <Navigate to="/" />;
  }

  return children;
};

export default ProtectedRoute;
