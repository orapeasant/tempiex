import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWorkflows } from '@/hooks/useWorkflows';
import { StatusBadge } from '@/components/StatusBadge';
import { format } from 'date-fns';

const STATUS_OPTIONS = ['', 'Running', 'Completed', 'Failed', 'Cancelled', 'TimedOut', 'Terminated'];

export function WorkflowList() {
  const { namespace = '' } = useParams<{ namespace: string }>();
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState('');
  const { data, isLoading, error } = useWorkflows(namespace, statusFilter);

  if (isLoading) return <div className="text-gray-500">Loading workflows...</div>;
  if (error) return <div className="text-red-500">Error: {(error as Error).message}</div>;

  const executions = data?.executions || [];

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-2">Workflows</h1>
      <p className="text-sm text-gray-500 mb-4">Namespace: <span className="font-medium text-blue-600">{namespace}</span></p>

      <div className="flex gap-3 mb-4">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          <option value="">All Statuses</option>
          {STATUS_OPTIONS.filter(Boolean).map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-800 text-white">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Workflow ID</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Type</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Task Queue</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Start Time</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">End Time</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {executions.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-400">No workflows found</td></tr>
            ) : (
              executions.map((wf) => (
                <tr
                  key={`${wf.execution.workflowId}-${wf.execution.runId}`}
                  onClick={() => navigate(`/namespaces/${namespace}/workflows/${wf.execution.workflowId}/${wf.execution.runId}`)}
                  className="hover:bg-blue-50 cursor-pointer transition-colors"
                >
                  <td className="px-4 py-3 text-sm font-medium text-blue-600">{wf.execution.workflowId}</td>
                  <td className="px-4 py-3 text-sm text-gray-700">{wf.type.name}</td>
                  <td className="px-4 py-3"><StatusBadge status={wf.status} /></td>
                  <td className="px-4 py-3 text-sm text-gray-600">{wf.taskQueue}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
                    {format(new Date(wf.startTime), 'yyyy-MM-dd HH:mm:ss')}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
                    {wf.closeTime ? format(new Date(wf.closeTime), 'yyyy-MM-dd HH:mm:ss') : '—'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
