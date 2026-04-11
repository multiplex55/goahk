import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import StatusBar from './StatusBar';

describe('StatusBar', () => {
  afterEach(() => cleanup());

  it('copies path and selector and shows notifications', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });

    render(<StatusBar statusText="Ready" errorText="" path="Desktop > App" selector={'name="App"'} />);

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

  it('shows error text when stage failure is preferred', () => {
    render(<StatusBar statusText="Loaded node details" errorText="Failed GetNodeDetails: boom" preferStageFailure path="Desktop" selector="" />);
    expect(screen.getByText('Failed GetNodeDetails: boom')).toBeInTheDocument();
  });

  it('shows no-selector feedback only when details are loaded and selector is absent', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    const onCopySelector = vi.fn().mockResolvedValue({ selector: '', clipboardUpdated: false });
    render(<StatusBar statusText="Loaded node details" errorText="" path="Desktop" selector="" hasDetails onCopySelector={onCopySelector} />);

    await act(async () => {
      const copyButtons = screen.getAllByRole('button', { name: 'Copy selector' });
      fireEvent.click(copyButtons[copyButtons.length - 1]);
    });

    expect(screen.getByRole('status')).toHaveTextContent('No selector available');
    expect(writeText).not.toHaveBeenCalled();
  });
});
