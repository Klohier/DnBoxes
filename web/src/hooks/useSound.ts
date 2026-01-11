import { useCallback, useRef, useEffect } from "react";

export const useSound = (soundFile: string) => {
  const audioRef = useRef<HTMLAudioElement | null>(null);

  useEffect(() => {
    audioRef.current = new Audio(soundFile);
    audioRef.current.preload = "auto";
    audioRef.current.load();

    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }
    };
  }, [soundFile]);

  const play = useCallback(() => {
    if (!audioRef.current) {
      audioRef.current = new Audio(soundFile);
      audioRef.current.preload = "auto";
    }

    audioRef.current.currentTime = 0;
    audioRef.current.play().catch((err) => {
      console.error("Error playing sound:", err);
    });
  }, [soundFile]);

  return play;
};
