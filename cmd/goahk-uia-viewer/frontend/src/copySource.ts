export function propertyCopySource(name: string, value: string): string {
  return `${name}=${value}`;
}

export function statusCopySource(value: string): string {
  return value;
}
