export function parseJsonInput(input: string): unknown {
  const trimmed = input.trim();
  if (!trimmed) return undefined;
  try {
    return JSON.parse(trimmed);
  } catch (e) {
    throw new Error(`Invalid JSON input: ${(e as Error).message}`);
  }
}

export function encodePayload(data: unknown): { data: string; metadata: { encoding: string } } {
  const json = JSON.stringify(data);
  const encoded = Buffer.from(json).toString('base64');
  return {
    data: encoded,
    metadata: { encoding: Buffer.from('json/plain').toString('base64') },
  };
}

export function decodePayload(payload: { data?: string }): unknown {
  if (!payload.data) return undefined;
  try {
    const json = Buffer.from(payload.data, 'base64').toString('utf-8');
    return JSON.parse(json);
  } catch {
    return payload.data;
  }
}

export function parseKeyValuePairs(pairs: string[]): Record<string, string> {
  const result: Record<string, string> = {};
  for (const pair of pairs) {
    const idx = pair.indexOf('=');
    if (idx === -1) throw new Error(`Invalid KEY=VALUE pair: ${pair}`);
    result[pair.slice(0, idx)] = pair.slice(idx + 1);
  }
  return result;
}
