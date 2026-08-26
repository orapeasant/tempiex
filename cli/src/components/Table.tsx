import React from 'react';
import { Box, Text } from 'ink';

export interface Column {
  key: string;
  header: string;
  width?: number;
  minWidth?: number;
}

interface TableProps {
  columns: Column[];
  rows: Record<string, unknown>[];
}

function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen - 1) + '…';
}

function cellStr(val: unknown): string {
  if (val === null || val === undefined) return '—';
  return String(val);
}

export function Table({ columns, rows }: TableProps): React.ReactElement {
  const widths = columns.map((col) => {
    const headerLen = col.header.length;
    const maxDataLen = rows.reduce((max, row) => {
      const cell = cellStr(row[col.key]);
      return Math.max(max, cell.length);
    }, 0);
    const computed = Math.max(headerLen, maxDataLen, col.minWidth ?? 0);
    return col.width ?? Math.min(computed, 40);
  });

  return (
    <Box flexDirection="column">
      <Box>
        {columns.map((col, i) => (
          <Box key={col.key} width={widths[i]! + 2}>
            <Text bold>{truncate(col.header, widths[i]!)}</Text>
          </Box>
        ))}
      </Box>
      <Box>
        {columns.map((col, i) => (
          <Box key={col.key} width={widths[i]! + 2}>
            <Text dimColor>{'-'.repeat(widths[i]!)}</Text>
          </Box>
        ))}
      </Box>
      {rows.map((row, ri) => (
        <Box key={ri}>
          {columns.map((col, i) => (
            <Box key={col.key} width={widths[i]! + 2}>
              <Text>{truncate(cellStr(row[col.key]), widths[i]!)}</Text>
            </Box>
          ))}
        </Box>
      ))}
      {rows.length === 0 && (
        <Box>
          <Text dimColor>No results.</Text>
        </Box>
      )}
    </Box>
  );
}
