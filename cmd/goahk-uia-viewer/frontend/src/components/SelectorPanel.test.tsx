import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import SelectorPanel from './SelectorPanel';

describe('SelectorPanel', () => {
  it('renders alternate selectors with rationale and supports copy actions', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    const onNotify = vi.fn();
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });

    render(
      <SelectorPanel
        selector="automationId=SearchBox"
        selectorPath={{
          fullPath: [
            { nodeID: 'node:win', localizedControlType: 'window' },
            { nodeID: 'node:pane', localizedControlType: 'pane' },
            { nodeID: 'node:edit', localizedControlType: 'edit' }
          ]
        }}
        selectorOptions={{
          best: {
            rank: 1,
            selector: { automationId: 'SearchBox' },
            rationale: 'automation id is stable',
            score: 100,
            source: 'automationId'
          },
          alternates: [
            {
              rank: 2,
              selector: { automationId: 'SearchBox', controlType: 'Edit' },
              rationale: 'narrows duplicate ids',
              score: 95,
              source: 'automationId+controlType'
            }
          ]
        }}
        onNotify={onNotify}
      />
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy Selector' }));
    });

    expect(writeText).toHaveBeenCalledWith('automationId=SearchBox');
    expect(screen.getByText('Rationale: automation id is stable')).toBeInTheDocument();
    expect(screen.getByText('Path: window > pane > edit')).toBeInTheDocument();
    expect(screen.getByText(/score: 95/)).toBeInTheDocument();

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy' }));
    });

    expect(writeText).toHaveBeenCalledWith('automationId=SearchBox;controlType=Edit');
    expect(onNotify).toHaveBeenCalledWith('Selector #2 copied', 'success');
  });
});
