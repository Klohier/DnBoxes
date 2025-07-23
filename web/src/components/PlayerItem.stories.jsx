import PlayerItem from "./PlayerItem";
import { UserContext } from "../UserContext";
import { AuthProvider } from "../AuthContext";
import { MemoryRouter } from "react-router-dom";

const meta = {
  component: PlayerItem,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <AuthProvider>
          <UserContext.Provider
            value={{
              user: {
                user_id: 123,
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
    player: {
      user_id: 456,
      username: "OtherPlayer",
      avatarUrl: null,
      status: "online",
    },
    onClick: (player) => console.log("Clicked", player),
  },
};
