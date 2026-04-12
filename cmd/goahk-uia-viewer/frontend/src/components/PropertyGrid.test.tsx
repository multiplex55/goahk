import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import PropertyGrid from './PropertyGrid';

describe('PropertyGrid', () => {
  const denseProperties = [
    { name: 'RuntimeID', value: '1.2.3', group: 'identity', status: 'ok' },
    { name: 'AutomationId', value: 'MainWindow', group: 'identity', status: 'ok' },
    { name: 'ProcessId', value: '4242', group: 'identity', status: 'ok' },
    { name: 'ClassName', value: 'CalcFrame', group: 'identity', status: 'ok' },
    { name: 'FrameworkId', value: 'Win32', group: 'identity', status: 'ok' },
    { name: 'ControlType', value: 'Window', group: 'semantics', status: 'ok' },
    { name: 'LocalizedControlType', value: 'window', group: 'semantics', status: 'ok' },
    { name: 'Name', value: 'Calculator', group: 'semantics', status: 'ok' },
    { name: 'Value', value: null, group: 'semantics', status: 'ok' },
    { name: 'HelpText', value: null, group: 'semantics', status: 'unsupported' },
    { name: 'ItemType', value: null, group: 'semantics', status: 'ok' },
    { name: 'ItemStatus', value: null, group: 'semantics', status: 'ok' },
    { name: 'AccessKey', value: null, group: 'semantics', status: 'unsupported' },
    { name: 'AcceleratorKey', value: null, group: 'semantics', status: 'ok' },
    { name: 'IsEnabled', value: 'true', group: 'state', status: 'ok' },
    { name: 'IsPassword', value: null, group: 'state', status: 'unsupported' },
    { name: 'IsOffscreen', value: 'false', group: 'state', status: 'ok' },
    { name: 'HasKeyboardFocus', value: 'false', group: 'state', status: 'ok' },
    { name: 'IsKeyboardFocusable', value: 'true', group: 'state', status: 'ok' },
    { name: 'IsRequiredForForm', value: null, group: 'state', status: 'unsupported' },
    { name: 'BoundingRectangle', value: '1,2,300,200', group: 'geometry', status: 'ok' },
    { name: 'LabeledBy', value: null, group: 'relation', status: 'ok' }
  ] as const;

  it('highlights selected rows and copies deterministic source', () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText }
    });

    render(
      <PropertyGrid
        element={{ name: 'Calculator', controlType: 'Window', isEnabled: true, isKeyboardFocusable: false, hasKeyboardFocus: false, isPassword: false, isOffscreen: false, isRequiredForForm: false }}
        properties={[
          { name: 'AutomationId', value: 'MainWindow', group: 'identity', status: 'ok' },
          { name: 'ControlType', value: 'Window', group: 'semantics', status: 'ok' }
        ]}
      />
    );

    const controlType = screen.getByText('ControlType').closest('.property-row');
    fireEvent.click(controlType!);
    expect(controlType).toHaveClass('selected');

    fireEvent.click(screen.getByRole('button', { name: 'Copy ControlType' }));
    expect(writeText).toHaveBeenCalledWith('ControlType=Window');
    expect(screen.getByText('Calculator (Window)')).toBeInTheDocument();
  });

  it('renders groups in stable order and displays explicit null/unsupported values', () => {
    render(
      <PropertyGrid
        properties={denseProperties.map((property) => ({ ...property }))}
        element={{ name: 'Calculator', controlType: 'Window', isEnabled: true, isKeyboardFocusable: true, hasKeyboardFocus: false, isPassword: false, isOffscreen: false, isRequiredForForm: false }}
      />
    );

    const headings = screen.getAllByRole('heading', { level: 4 }).map((heading) => heading.textContent);
    expect(headings).toEqual(['identity', 'semantics', 'state', 'geometry', 'relation']);
    expect(screen.getByText('null')).toBeInTheDocument();
    expect(screen.getAllByText('unsupported').length).toBeGreaterThan(0);
  });

  it('renders the full dense property set', () => {
    render(
      <PropertyGrid
        properties={denseProperties.map((property) => ({ ...property }))}
        element={{ name: 'Calculator', controlType: 'Window', isEnabled: true, isKeyboardFocusable: true, hasKeyboardFocus: false, isPassword: false, isOffscreen: false, isRequiredForForm: false }}
      />
    );

    expect(screen.getAllByRole('button', { name: /Copy / }).length).toBe(22);
    expect(screen.getByText('BoundingRectangle')).toBeInTheDocument();
    expect(screen.getByText('LabeledBy')).toBeInTheDocument();
  });
});
