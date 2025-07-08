import Chatbox from "../components/ChatBox";
import { useAuth } from "../AuthContext";
// import PlayerList from "../components/PlayerList";
import PlayerLobby from "../components/PlayerLobby";

const Home = () => {
  const { logout } = useAuth();

  return (
    <>
      <PlayerLobby></PlayerLobby>
      <Chatbox sessionID={1}></Chatbox>
      <button onClick={logout} style={{ marginTop: "20px" }}>
        Logout
      </button>
    </>
  );
};

export default Home;
