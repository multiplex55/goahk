import { useState } from 'react';
import { propertyCopySource } from '../copySource';
import { PropertyItem } from '../types';

export type NotifyFn = (message: string, variant?: 'success' | 'error') => void;

type PropertyGridProps = {
  properties: PropertyItem[];
  onNotify?: NotifyFn;
};

export default function PropertyGrid({ properties, onNotify }: PropertyGridProps) {
  const [selectedName, setSelectedName] = useState<string | null>(null);

  async function copyProperty(property: PropertyItem) {
    try {
      const source = propertyCopySource(property.name, property.value);
      await navigator.clipboard.writeText(source);
      onNotify?.(`Copied ${property.name}`, 'success');
    } catch {
      onNotify?.(`Failed to copy ${property.name}`, 'error');
    }
  }

  return (
    <section aria-label="property grid">
      <h3>Properties</h3>
      <dl className="property-grid">
        {properties.map((property) => (
          <div
            key={property.name}
            className={`property-row ${selectedName === property.name ? 'selected' : ''}`.trim()}
            onClick={() => setSelectedName(property.name)}
          >
            <dt>{property.name}</dt>
            <dd>{property.value}</dd>
            <button
              type="button"
              className="copy-button"
              aria-label={`Copy ${property.name}`}
              onClick={() => void copyProperty(property)}
            >
              ⧉
            </button>
          </div>
        ))}
      </dl>
    </section>
  );
}
