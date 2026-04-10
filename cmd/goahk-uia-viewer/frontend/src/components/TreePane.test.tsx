import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import TreePane from './TreePane';

describe('TreePane', () => {
  it('renders lazy tree nodes and toggles expandable nodes', () => {
    const onToggleNode = vi.fn();

    render(
      <TreePane
        rootNodes={[
          { id: 'root', name: 'Desktop', hasChildren: true },
          { id: 'leaf', name: 'Leaf', hasChildren: false }
        ]}
        expandedNodeIds={new Set(['root'])}
        onToggleNode={onToggleNode}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /Desktop/ }));
    fireEvent.click(screen.getByRole('button', { name: /Leaf/ }));

    expect(onToggleNode).toHaveBeenCalledTimes(1);
    expect(onToggleNode).toHaveBeenCalledWith('root');
  });
});
