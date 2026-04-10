import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import WindowInfoPanel from './WindowInfoPanel';

describe('WindowInfoPanel', () => {
  it('renders window info, properties, patterns, and selector sections', () => {
    render(
      <WindowInfoPanel
        windowTitle="Calculator"
        properties={[{ name: 'Name', value: 'Calculator' }]}
        patternActions={[{ id: 'invoke', label: 'Invoke', supported: true }]}
        selectorText="name=Calculator"
        onInvokePattern={vi.fn()}
      />
    );

    expect(screen.getByRole('heading', { name: 'Window Info' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Properties' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Patterns' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Best Selector' })).toBeInTheDocument();
  });
});
