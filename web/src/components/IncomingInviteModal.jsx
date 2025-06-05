/* eslint-disable react/prop-types */
import styles from "./IncomingInviteModal.module.css";
const IncomingInviteModal = ({
  incomingInvite,
  onAccept,
  onDecline,
  onClose,
}) => {
  if (!incomingInvite) return null;

  return (
    <>
      <div className={styles.modal}>
        <h4>Game Invite</h4>
        <p>
          {incomingInvite.senderName} has invited you to a game with a board
          size of {incomingInvite.board_size}.
        </p>
        <button onClick={onAccept}>Accept</button>
        <button onClick={onDecline}>Decline</button>
      </div>

      <div onClick={onClose} className={styles.overlay} />
    </>
  );
};

export default IncomingInviteModal;
