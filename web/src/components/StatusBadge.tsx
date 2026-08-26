interface StatusBadgeProps {
  status: string;
}

const statusColors: Record<string, string> = {
  Running: 'bg-blue-500 text-white',
  Completed: 'bg-green-500 text-white',
  Failed: 'bg-red-500 text-white',
  Cancelled: 'bg-orange-500 text-white',
  TimedOut: 'bg-yellow-500 text-black',
  ContinuedAsNew: 'bg-purple-500 text-white',
  Terminated: 'bg-red-900 text-white',
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const color = statusColors[status] || 'bg-gray-400 text-white';
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${color}`}>
      {status}
    </span>
  );
}
