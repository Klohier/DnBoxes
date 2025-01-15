import Chatbox from "../components/ChatBox";
import { useAuth } from "../AuthContext";
import PlayerList from "../components/PlayerList";

const Home = () => {
  const { logout } = useAuth();

  return (
    <>
      <PlayerList></PlayerList>
      <Chatbox></Chatbox>
      <button onClick={logout} style={{ marginTop: "20px" }}>
        Logout
      </button>
    </>
  );
};

export default Home;
