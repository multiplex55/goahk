import { useState } from 'react';
import { propertyCopySource } from '../copySource';
import { ElementDetails, PropertyItem } from '../types';

export type NotifyFn = (message: string, variant?: 'success' | 'error') => void;

type PropertyGridProps = {
  properties: PropertyItem[];
  element?: ElementDetails;
  onNotify?: NotifyFn;
};

export default function PropertyGrid({ properties, element, onNotify }: PropertyGridProps) {
  const [selectedName, setSelectedName] = useState<string | null>(null);
  const orderedGroups: PropertyItem['group'][] = ['identity', 'semantics', 'state', 'geometry', 'relation'];
  const propertiesByGroup = orderedGroups.map((group) => ({
    group,
    rows: properties.filter((property) => property.group === group)
  }));

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
      {element?.name ? <p className="window-info">{element.name} ({element.controlType ?? 'Unknown'})</p> : null}
      {propertiesByGroup.map(({ group, rows }) => (
        <div key={group}>
          <h4>{group}</h4>
          <dl className="property-grid">
            {rows.map((property) => (
              <div
                key={property.name}
                className={`property-row ${selectedName === property.name ? 'selected' : ''}`.trim()}
                onClick={() => setSelectedName(property.name)}
              >
                <dt>{property.name}</dt>
                <dd>{property.status === 'unsupported' ? 'unsupported' : property.value ?? 'null'}</dd>
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
        </div>
      ))}
    </section>
  );
}
