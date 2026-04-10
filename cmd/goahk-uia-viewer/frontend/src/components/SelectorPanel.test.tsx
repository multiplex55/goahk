import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import SelectorPanel from './SelectorPanel';

describe('SelectorPanel', () => {
  it('copies selector text when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    const onNotify = vi.fn();
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });

    render(<SelectorPanel selector="type=Button" onNotify={onNotify} />);

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy Selector' }));
    });

    expect(writeText).toHaveBeenCalledWith('type=Button');
    expect(onNotify).toHaveBeenCalledWith('Selector copied', 'success');
  });
});
