import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import TreePane from './TreePane';

afterEach(() => cleanup());

describe('TreePane', () => {
  it('renders recursively nested nodes when expanded', () => {
    render(
      <TreePane
        nodesByID={{
          root: { id: 'root', name: 'Desktop', hasChildren: true },
          child: { id: 'child', name: 'Child', hasChildren: true },
          leaf: { id: 'leaf', name: 'Leaf', hasChildren: false }
        }}
        childrenByParentID={{
          __root__: ['root'],
          root: ['child'],
          child: ['leaf']
        }}
        expandedByID={{ root: true, child: true }}
        onToggleNode={vi.fn()}
      />
    );

    expect(screen.getByRole('button', { name: 'Select Desktop' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Select Child' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Select Leaf' })).toBeInTheDocument();
  });

  it('chevron toggles expansion only', () => {
    const onToggleNode = vi.fn();
    const onSelectNode = vi.fn();

    render(
      <TreePane
        nodesByID={{ root: { id: 'root', name: 'Desktop', hasChildren: true } }}
        childrenByParentID={{ __root__: ['root'] }}
        expandedByID={{}}
        onToggleNode={onToggleNode}
        onSelectNode={onSelectNode}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /Expand Desktop/i }));

    expect(onToggleNode).toHaveBeenCalledWith('root');
    expect(onSelectNode).not.toHaveBeenCalled();
  });

  it('row click selects without triggering toggle', () => {
    const onToggleNode = vi.fn();
    const onSelectNode = vi.fn();

    render(
      <TreePane
        nodesByID={{ root: { id: 'root', name: 'Desktop', hasChildren: true } }}
        childrenByParentID={{ __root__: ['root'] }}
        expandedByID={{}}
        onToggleNode={onToggleNode}
        onSelectNode={onSelectNode}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Select Desktop' }));

    expect(onSelectNode).toHaveBeenCalledWith('root');
    expect(onToggleNode).not.toHaveBeenCalled();
  });
});
