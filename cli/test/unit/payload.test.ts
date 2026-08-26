import { describe, it, expect } from 'vitest';
import { parseJsonInput, encodePayload, decodePayload, parseKeyValuePairs } from '../../src/utils/payload.js';

describe('parseJsonInput', () => {
  it('parses valid JSON objects', () => {
    const result = parseJsonInput('{"orderId": "123"}');
    expect(result).toEqual({ orderId: '123' });
  });

  it('parses JSON arrays', () => {
    const result = parseJsonInput('[1,2,3]');
    expect(result).toEqual([1, 2, 3]);
  });

  it('parses JSON strings', () => {
    const result = parseJsonInput('"hello"');
    expect(result).toBe('hello');
  });

  it('returns undefined for empty string', () => {
    const result = parseJsonInput('');
    expect(result).toBeUndefined();
  });

  it('throws for invalid JSON', () => {
    expect(() => parseJsonInput('{invalid}')).toThrow('Invalid JSON input');
  });
});

describe('encodePayload', () => {
  it('encodes data as base64 JSON', () => {
    const result = encodePayload({ foo: 'bar' });
    expect(result.data).toBeDefined();
    const decoded = Buffer.from(result.data, 'base64').toString('utf-8');
    expect(JSON.parse(decoded)).toEqual({ foo: 'bar' });
  });

  it('includes encoding metadata', () => {
    const result = encodePayload({});
    const encoding = Buffer.from(result.metadata.encoding, 'base64').toString('utf-8');
    expect(encoding).toBe('json/plain');
  });
});

describe('decodePayload', () => {
  it('decodes base64 JSON payload', () => {
    const data = Buffer.from(JSON.stringify({ x: 42 })).toString('base64');
    const result = decodePayload({ data });
    expect(result).toEqual({ x: 42 });
  });

  it('returns undefined for missing data', () => {
    const result = decodePayload({});
    expect(result).toBeUndefined();
  });
});

describe('parseKeyValuePairs', () => {
  it('parses KEY=VALUE pairs', () => {
    const result = parseKeyValuePairs(['foo=bar', 'hello=world']);
    expect(result).toEqual({ foo: 'bar', hello: 'world' });
  });

  it('handles value with equals sign', () => {
    const result = parseKeyValuePairs(['url=http://example.com?a=1']);
    expect(result).toEqual({ url: 'http://example.com?a=1' });
  });

  it('throws for invalid pair', () => {
    expect(() => parseKeyValuePairs(['noequalssign'])).toThrow('Invalid KEY=VALUE pair');
  });
});
