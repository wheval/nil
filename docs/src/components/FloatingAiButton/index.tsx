import React from 'react';
import { useColorMode } from '@docusaurus/theme-common';
import LightAiIcon from '../../../static/img/lightAiIcon.json';
import DarkAiIcon from '../../../static/img/darkAiIcon.json';
import styles from './styles.module.css';
import Modal from "react-modal";
import { AIChat } from "../AIChat";

const Lottie =
  typeof window !== 'undefined' ? require('lottie-react').default : () => null;

export default function FloatingAiButton() {
  const { colorMode } = useColorMode();
  const [modalIsOpen, setIsOpen] = React.useState(false);

  function toggleModal() {
    setIsOpen(prevState => !prevState);
  }

  return (
    <>
      <button className={styles.floatingAiButton} onClick={toggleModal}>
        <Lottie
          animationData={colorMode === 'dark' ? DarkAiIcon : LightAiIcon}
          loop={true}
          style={{ height: 30, width: 30 }}
        />
        Ask = nil; AI
      </button>

      <Modal
        isOpen={modalIsOpen}
        onRequestClose={toggleModal}
        contentLabel="Ask AI"
      >
        <div className={styles.closeModalButton} onClick={toggleModal}>
          Close Chat
        </div>
        <AIChat />
      </Modal>
    </>
  );
}