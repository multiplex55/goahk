import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import SelectorPanel from './SelectorPanel';

describe('SelectorPanel', () => {
  it('copies selector text when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    const onNotify = vi.fn();
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });

    render(
      <SelectorPanel
        selector="type=Button"
        selectorPath={{
          fullPath: [{ nodeID: 'node:root', name: 'Root' }, { nodeID: 'node:child', name: 'Child' }],
          selectorSuggestions: [{ rank: 1, selector: { automationId: 'ok' }, source: 'automationId' }]
        }}
        onNotify={onNotify}
      />
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy Selector' }));
    });

    expect(writeText).toHaveBeenCalledWith('type=Button');
    expect(onNotify).toHaveBeenCalledWith('Selector copied', 'success');
    expect(screen.getByText('Path: Root > Child')).toBeInTheDocument();
    expect(screen.getByText('#1 automationId')).toBeInTheDocument();
  });
});
