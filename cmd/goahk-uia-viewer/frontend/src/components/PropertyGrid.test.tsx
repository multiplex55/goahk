import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import PropertyGrid from './PropertyGrid';

describe('PropertyGrid', () => {
  it('highlights selected rows and copies deterministic source', () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText }
    });

    render(
      <PropertyGrid
        properties={[
          { name: 'AutomationId', value: 'MainWindow' },
          { name: 'ControlType', value: 'Window' }
        ]}
      />
    );

    const controlType = screen.getByText('ControlType').closest('.property-row');
    fireEvent.click(controlType!);
    expect(controlType).toHaveClass('selected');

    fireEvent.click(screen.getByRole('button', { name: 'Copy ControlType' }));
    expect(writeText).toHaveBeenCalledWith('ControlType=Window');
  });
});
