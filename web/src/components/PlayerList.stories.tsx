import PlayerList from "./PlayerList";
import { MemoryRouter } from "react-router-dom";
import { AuthProvider } from "../AuthContext";
import { UserContext } from "../UserContext";
import { Player } from "@/types/websocket";

const mockPlayers: Player[] = [
  {
    user_id: 123,
    username: "testuser",
    avatarUrl: "",
    status: "online",
  },
  {
    user_id: 456,
    username: "otherplayer1",
    avatarUrl: "",
    status: "offline",
  },
  {
    user_id: 789,
    username: "otherplayer2",
    avatarUrl: "",
    status: "online",
  },
];

const meta = {
  component: PlayerList,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <AuthProvider>
          <UserContext.Provider
            value={{
              user: {
                userID: 123, // current user ID
                username: "testuser",
              },
            }}
          >
            <Story />
          </UserContext.Provider>
        </AuthProvider>
      </MemoryRouter>
    ),
  ],
};

export default meta;

export const Default = {
  args: {
    players: mockPlayers,
    onPlayerClick: (player: Player) => {
      console.log("Player clicked:", player.username);
    },
  },
};
