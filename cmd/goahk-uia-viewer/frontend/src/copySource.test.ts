import { describe, expect, it } from 'vitest';
import { propertyCopySource, statusCopySource } from './copySource';

describe('copySource', () => {
  it('formats property strings deterministically', () => {
    expect(propertyCopySource('Name', 'Submit')).toBe('Name=Submit');
  });

  it('returns status strings unchanged', () => {
    expect(statusCopySource('Desktop > App')).toBe('Desktop > App');
  });
});
