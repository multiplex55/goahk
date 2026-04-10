import { useState } from 'react';
import { propertyCopySource } from '../copySource';
import { PropertyItem } from '../types';

type PropertyGridProps = {
  properties: PropertyItem[];
};

export default function PropertyGrid({ properties }: PropertyGridProps) {
  const [selectedName, setSelectedName] = useState<string | null>(null);

  async function copyProperty(property: PropertyItem) {
    const source = propertyCopySource(property.name, property.value);
    await navigator.clipboard.writeText(source);
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
              onClick={() => copyProperty(property)}
            >
              ⧉
            </button>
          </div>
        ))}
      </dl>
    </section>
  );
}
