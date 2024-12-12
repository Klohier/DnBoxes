import Chatbox from "../components/ChatBox";
// import Grid from "../components/SvgGrid";
import { useNavigate } from "react-router-dom";
import Cookies from "js-cookie";
import { useUser } from "../UserContext"; // Import the useUser hook
import PlayerList from "../components/PlayerList";

const Home = () => {
  const navigate = useNavigate();
  const { logoutUser } = useUser();

  const handleLogout = () => {
    // Remove the session cookie
    Cookies.remove("DnB-Session");
    logoutUser();
    // Redirect to the login page
    navigate("/");
  };
  return (
    <>
      <PlayerList></PlayerList>
      <Chatbox></Chatbox>
      <button onClick={handleLogout} style={{ marginTop: "20px" }}>
        Logout
      </button>
    </>
  );
};

export default Home;
