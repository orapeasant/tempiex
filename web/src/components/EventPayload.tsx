import { useState } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';

interface EventPayloadProps {
  attributes: Record<string, unknown>;
}

export function EventPayload({ attributes }: EventPayloadProps) {
  const [expanded, setExpanded] = useState(false);
  const hasContent = Object.keys(attributes).length > 0;

  if (!hasContent) return <span className="text-gray-400 text-xs">—</span>;

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800"
      >
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        {expanded ? 'Hide' : 'Show'} payload
      </button>
      {expanded && (
        <pre className="mt-1 p-2 bg-gray-50 rounded text-xs overflow-auto max-h-64 border">
          {JSON.stringify(attributes, null, 2)}
        </pre>
      )}
    </div>
  );
}
