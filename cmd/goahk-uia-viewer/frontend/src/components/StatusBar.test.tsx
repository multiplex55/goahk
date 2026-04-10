import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import StatusBar from './StatusBar';

describe('StatusBar', () => {
  it('copies path and selector and shows notifications', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });

    render(<StatusBar status="Ready" path="Desktop > App" selector={'name="App"'} />);

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Desktop > App' }));
    });
    expect(writeText).toHaveBeenCalledWith('Desktop > App');
    expect(screen.getByRole('status')).toHaveTextContent('Path copied');

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy selector' }));
    });
    expect(writeText).toHaveBeenCalledWith('name="App"');
    expect(screen.getByRole('status')).toHaveTextContent('Selector copied');
  });
});
