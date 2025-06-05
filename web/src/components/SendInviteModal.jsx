/* eslint-disable react/prop-types */
// eslint-disable-next-line react/prop-types
import styles from "./SendInviteModal.module.css";
const SendInviteModal = ({
  selectedPlayer,
  boardSize,
  onBoardSizeChange,
  onSendInvite,
  onClose,
}) => {
  if (!selectedPlayer) return null;

  return (
    <>
      <div className={styles.modal}>
        <h4>Send Game Invite</h4>
        <p>
          Are you sure you want to send a game invite to{" "}
          {selectedPlayer.username}?
        </p>
        <div>
          <label htmlFor="boardSize">Board Size: </label>
          <input
            type="number"
            id="boardSize"
            value={boardSize}
            onChange={(e) => onBoardSizeChange(parseInt(e.target.value))}
            min="1"
            max="100"
            style={{ marginLeft: "10px", width: "50px" }}
          />
        </div>
        <button onClick={onSendInvite}>Send Invite</button>
        <button onClick={onClose}>Cancel</button>
      </div>

      <div onClick={onClose} className={styles.overlay} />
    </>
  );
};

export default SendInviteModal;
