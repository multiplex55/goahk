import { ReactNode, useMemo } from 'react';
import Splitter from '../components/Splitter';

type ThreeColumnLayoutProps = {
  leftWidthPx: number;
  middleWidthPx: number;
  minPaneWidthPx?: number;
  maxPaneWidthPx?: number;
  onResize: (next: { leftWidthPx: number; middleWidthPx: number }) => void;
  left: ReactNode;
  middle: ReactNode;
  right: ReactNode;
};

const SPLITTER_WIDTH = 6;

export default function ThreeColumnLayout({
  leftWidthPx,
  middleWidthPx,
  minPaneWidthPx = 220,
  maxPaneWidthPx = 720,
  onResize,
  left,
  middle,
  right
}: ThreeColumnLayoutProps) {
  const boundedLeft = Math.min(maxPaneWidthPx, Math.max(minPaneWidthPx, leftWidthPx));
  const boundedMiddle = Math.min(maxPaneWidthPx, Math.max(minPaneWidthPx, middleWidthPx));

  const style = useMemo(
    () => ({
      gridTemplateColumns: `${boundedLeft}px ${SPLITTER_WIDTH}px ${boundedMiddle}px ${SPLITTER_WIDTH}px minmax(260px, 1fr)`
    }),
    [boundedLeft, boundedMiddle]
  );

  const dragLeft = (deltaPx: number) => {
    onResize({
      leftWidthPx: Math.min(maxPaneWidthPx, Math.max(minPaneWidthPx, boundedLeft + deltaPx)),
      middleWidthPx: boundedMiddle
    });
  };

  const dragMiddle = (deltaPx: number) => {
    onResize({
      leftWidthPx: boundedLeft,
      middleWidthPx: Math.min(maxPaneWidthPx, Math.max(minPaneWidthPx, boundedMiddle + deltaPx))
    });
  };

  return (
    <main className="three-column-layout" style={style}>
      {left}
      <Splitter
        ariaLabel="Resize window list pane"
        valueMin={minPaneWidthPx}
        valueMax={maxPaneWidthPx}
        valueNow={boundedLeft}
        onDrag={dragLeft}
      />
      {middle}
      <Splitter
        ariaLabel="Resize details pane"
        valueMin={minPaneWidthPx}
        valueMax={maxPaneWidthPx}
        valueNow={boundedMiddle}
        onDrag={dragMiddle}
      />
      {right}
    </main>
  );
}
