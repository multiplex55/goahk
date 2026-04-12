export function propertyCopySource(name: string, value: string | null): string {
  return `${name}=${value ?? 'null'}`;
}

export function statusCopySource(value: string): string {
  return value;
}
