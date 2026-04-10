import { useRef } from 'react';

type SplitterProps = {
  onDrag: (deltaPx: number) => void;
  ariaLabel: string;
};

export default function Splitter({ onDrag, ariaLabel }: SplitterProps) {
  const startXRef = useRef(0);
  const draggingRef = useRef(false);

  const onMouseDown = (event: React.MouseEvent<HTMLDivElement>) => {
    draggingRef.current = true;
    startXRef.current = event.clientX;
  };

  const onMouseMove = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!draggingRef.current) {
      return;
    }
    const delta = event.clientX - startXRef.current;
    if (delta === 0) {
      return;
    }
    startXRef.current = event.clientX;
    onDrag(delta);
  };

  const onMouseUp = () => {
    draggingRef.current = false;
  };

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      aria-label={ariaLabel}
      className="splitter"
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={onMouseUp}
      onMouseLeave={onMouseUp}
    />
  );
}
