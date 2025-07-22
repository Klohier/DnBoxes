/* eslint-disable react/prop-types */
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { useUser } from "../UserContext";
const PlayerItem = ({ player, onClick }) => {
  const { user } = useUser();
  const isCurrentUser = player.user_id === user?.userID;

  const handleClick = () => {
    if (!isCurrentUser) {
      onClick(player);
    }
  };
  return (
    <li
      onClick={handleClick}
      className={`flex justify-between items-center p-3 border rounded-md transition
        ${
          isCurrentUser
            ? "cursor-default bg-gray-100 text-gray-500"
            : "cursor-pointer hover:bg-muted"
        }
      `}
      title={isCurrentUser ? "This is you" : `Invite ${player.username}`}
    >
      <div className="flex items-center space-x-3">
        <Avatar>
          {player.avatarUrl ? (
            <AvatarImage src={player.avatarUrl} alt={player.username} />
          ) : (
            <AvatarFallback>{player.username[0]}</AvatarFallback>
          )}
        </Avatar>
        <div>
          <div className="font-medium">{player.username}</div>
          <div className="text-xs text-muted-foreground">
            ID: {player.user_id}
          </div>
        </div>
      </div>
      {player.status && (
        <Badge variant={player.status === "online" ? "success" : "default"}>
          {player.status}
        </Badge>
      )}
    </li>
  );
};

export default PlayerItem;
