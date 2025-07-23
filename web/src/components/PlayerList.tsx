import PlayerItem from "./PlayerItem";
import { Player } from "@/types/websocket";

interface PlayerListProps {
  players: Player[];
  onPlayerClick: (player: Player) => void;
}

const PlayerList: React.FC<PlayerListProps> = ({ players, onPlayerClick }) => {
  return (
    <ul>
      {players.map((player) => (
        <PlayerItem
          key={player.user_id}
          player={player}
          onClick={onPlayerClick}
        />
      ))}
    </ul>
  );
};

export default PlayerList;
