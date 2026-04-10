import { WindowItem } from '../types';

type WindowListPaneProps = {
  windows: WindowItem[];
  selectedWindowId?: string;
  onSelectWindow: (id: string) => void;
};

export default function WindowListPane({ windows, selectedWindowId, onSelectWindow }: WindowListPaneProps) {
  return (
    <section className="pane" aria-label="window list pane">
      <h2>Windows</h2>
      <ul className="list">
        {windows.map((window) => (
          <li key={window.id}>
            <button
              type="button"
              className={selectedWindowId === window.id ? 'selected' : ''}
              onClick={() => onSelectWindow(window.id)}
            >
              <span>{window.title || '<untitled>'}</span>
              <small>{window.processName}</small>
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
