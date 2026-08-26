import { describe, it, expect } from 'vitest';
import { renderJson, renderYaml, isNonTextOutput } from '../../src/utils/output.js';

describe('output utils', () => {
  it('renderJson produces valid JSON', () => {
    const data = { foo: 'bar', num: 42, arr: [1, 2, 3] };
    const out = renderJson(data);
    expect(JSON.parse(out)).toEqual(data);
    expect(out).toContain('"foo": "bar"');
  });

  it('renderYaml produces YAML string', () => {
    const data = { key: 'value', nested: { a: 1 } };
    const out = renderYaml(data);
    expect(out).toContain('key: value');
    expect(out).toContain('a: 1');
  });

  it('isNonTextOutput returns true for json and yaml', () => {
    expect(isNonTextOutput('json')).toBe(true);
    expect(isNonTextOutput('yaml')).toBe(true);
    expect(isNonTextOutput('text')).toBe(false);
  });

  it('renderJson handles arrays', () => {
    const data = [1, 2, 3];
    const out = renderJson(data);
    expect(JSON.parse(out)).toEqual([1, 2, 3]);
  });

  it('renderJson handles null', () => {
    const out = renderJson(null);
    expect(out).toBe('null');
  });
});
