/* eslint-disable react/prop-types */

import PlayerItem from "./PlayerItem";

const PlayerList = ({ players, onPlayerClick }) => {
  return (
    <ul>
      {players.map((player) => (
        <PlayerItem
          key={player.userID}
          player={player}
          onClick={onPlayerClick}
        />
      ))}
    </ul>
  );
};

export default PlayerList;
