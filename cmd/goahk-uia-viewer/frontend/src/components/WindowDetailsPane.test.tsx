import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import WindowDetailsPane from './WindowDetailsPane';

describe('WindowDetailsPane', () => {
  it('renders info, property grid, and pattern panel', () => {
    const onInvokePattern = vi.fn();

    render(
      <WindowDetailsPane
        windowTitle="Settings"
        properties={[{ name: 'Name', value: 'Settings' }]}
        patternActions={[{ id: 'invoke', label: 'Invoke' }]}
        onInvokePattern={onInvokePattern}
      />
    );

    expect(screen.getByText('Window Info')).toBeInTheDocument();
    expect(screen.getAllByText('Settings')).toHaveLength(2);
    expect(screen.getByText('Name')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Invoke' }));
    expect(onInvokePattern).toHaveBeenCalledWith('invoke');
  });
});
