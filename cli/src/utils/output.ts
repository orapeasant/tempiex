import { stringify as yamlStringify } from 'yaml';

export type OutputFormat = 'text' | 'json' | 'yaml';

export function renderJson(data: unknown): string {
  return JSON.stringify(data, null, 2);
}

export function renderYaml(data: unknown): string {
  return yamlStringify(data);
}

export function printOutput(data: unknown, format: OutputFormat): void {
  if (format === 'json') {
    console.log(renderJson(data));
  } else if (format === 'yaml') {
    console.log(renderYaml(data));
  }
  // 'text' is handled by Ink render
}

export function isNonTextOutput(format: OutputFormat): boolean {
  return format === 'json' || format === 'yaml';
}
