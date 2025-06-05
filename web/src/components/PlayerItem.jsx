/* eslint-disable react/prop-types */
import styles from "./PlayerItem.module.css";

const PlayerItem = ({ player, onClick }) => {
  return (
    <li onClick={() => onClick(player)} className={styles.playerItem}>
      {player.username} (ID: {player.userID})
    </li>
  );
};

export default PlayerItem;
