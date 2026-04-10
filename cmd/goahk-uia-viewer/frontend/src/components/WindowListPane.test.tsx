import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import WindowListPane from './WindowListPane';

describe('WindowListPane', () => {
  it('renders windows and dispatches selection', () => {
    const onSelectWindow = vi.fn();

    render(
      <WindowListPane
        windows={[
          { id: '1', title: 'Notepad', processName: 'notepad.exe' },
          { id: '2', title: 'Calculator', processName: 'calculator.exe' }
        ]}
        selectedWindowId="1"
        onSelectWindow={onSelectWindow}
      />
    );

    expect(screen.getByRole('heading', { name: 'Windows and Controls' })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /Calculator/ }));
    expect(onSelectWindow).toHaveBeenCalledWith('2');
  });
});
