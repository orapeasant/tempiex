import {
  formatDistanceToNow,
  formatDuration as fmtDuration,
  intervalToDuration,
} from 'date-fns';

export function formatRelative(dateStr: string | undefined): string {
  if (!dateStr) return '—';
  try {
    return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
  } catch {
    return dateStr;
  }
}

export function formatDuration(ms: number): string {
  const duration = intervalToDuration({ start: 0, end: ms });
  return (
    fmtDuration(duration, {
      format: ['hours', 'minutes', 'seconds'],
    }) || '0s'
  );
}

export function formatDateTime(dateStr: string | undefined): string {
  if (!dateStr) return '—';
  try {
    return new Date(dateStr).toISOString().replace('T', ' ').replace('Z', ' UTC');
  } catch {
    return dateStr;
  }
}
