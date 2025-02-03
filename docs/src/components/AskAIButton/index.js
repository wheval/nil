import styles from "./styles.module.css";
import React from "react";
import Modal from "react-modal";
import { AIChat } from "../AIChat";

const customStyles = {
  content: {
    top: "50%",
    left: "50%",
    right: "auto",
    bottom: "auto",
    margginTop: "-50%",
    borderRadius: "25px",
    marginRight: "-50%",
    transform: "translate(-50%, -50%)",
    background: "#000",
  },
};

export function AskAiButton() {
  const [modalIsOpen, setIsOpen] = React.useState(false);

  function openModal() {
    setIsOpen(true);
  }

  function closeModal() {
    setIsOpen(false);
  }

  return (
    <div>
      <div className={styles.askAIButton} onClick={openModal}>
        <span className={styles.buttonText}>Ask AI&nbsp;</span>
        <span class="material-icons">smart_toy</span>
      </div>
      <Modal
        isOpen={modalIsOpen}
        onRequestClose={closeModal}
        contentLabel="Ask AI"
        styles={customStyles}
      >
        <div className={styles.closeModalButton} onClick={closeModal}>
          Close Chat
        </div>
        {<AIChat />}
      </Modal>
    </div>
  );
}
