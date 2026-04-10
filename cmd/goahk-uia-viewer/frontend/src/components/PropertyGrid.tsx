import { PropertyItem } from '../types';

type PropertyGridProps = {
  properties: PropertyItem[];
};

export default function PropertyGrid({ properties }: PropertyGridProps) {
  return (
    <section aria-label="property grid">
      <h3>Properties</h3>
      <dl className="property-grid">
        {properties.map((property) => (
          <div key={property.name} className="property-row">
            <dt>{property.name}</dt>
            <dd>{property.value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
