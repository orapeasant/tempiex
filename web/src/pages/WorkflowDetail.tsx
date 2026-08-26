import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWorkflowDetail } from '@/hooks/useWorkflowDetail';
import { useEventHistory } from '@/hooks/useEventHistory';
import { StatusBadge } from '@/components/StatusBadge';
import { EventPayload } from '@/components/EventPayload';
import { ArrowLeft } from 'lucide-react';
import { format } from 'date-fns';

type TabName = 'history' | 'input';

export function WorkflowDetail() {
  const { namespace = '', workflowId = '', runId = '' } = useParams<{
    namespace: string;
    workflowId: string;
    runId: string;
  }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<TabName>('history');

  const { data: detailData, isLoading: detailLoading } = useWorkflowDetail(namespace, workflowId, runId);
  const { data: historyData, isLoading: historyLoading } = useEventHistory(namespace, workflowId, runId);

  if (detailLoading) return <div className="text-gray-500">Loading...</div>;

  const info = detailData?.workflowExecutionInfo;
  if (!info) return <div className="text-red-500">Workflow not found</div>;

  const events = historyData?.history?.events || [];

  return (
    <div>
      <button
        onClick={() => navigate(`/namespaces/${namespace}/workflows`)}
        className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 mb-4"
      >
        <ArrowLeft size={16} />
        Back to workflows
      </button>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center gap-3 mb-4">
          <h1 className="text-xl font-bold text-gray-900">{info.execution.workflowId}</h1>
          <StatusBadge status={info.status} />
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div><span className="text-gray-500">Run ID</span><p className="font-mono text-xs mt-1">{info.execution.runId}</p></div>
          <div><span className="text-gray-500">Type</span><p className="mt-1">{info.type.name}</p></div>
          <div><span className="text-gray-500">Task Queue</span><p className="mt-1">{info.taskQueue}</p></div>
          <div><span className="text-gray-500">Started</span><p className="mt-1">{format(new Date(info.startTime), 'yyyy-MM-dd HH:mm:ss')}</p></div>
          {info.closeTime && (
            <div><span className="text-gray-500">Closed</span><p className="mt-1">{format(new Date(info.closeTime), 'yyyy-MM-dd HH:mm:ss')}</p></div>
          )}
        </div>
      </div>

      <div className="flex gap-1 mb-4 border-b">
        <button
          onClick={() => setActiveTab('history')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'history' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
        >
          Event History ({events.length})
        </button>
        <button
          onClick={() => setActiveTab('input')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'input' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'}`}
        >
          Input &amp; Result
        </button>
      </div>

      {activeTab === 'history' && (
        historyLoading ? (
          <div className="text-gray-500">Loading events...</div>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-semibold uppercase w-16">ID</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Event Type</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Time</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Payload</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {events.map((evt) => (
                  <tr key={evt.eventId} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-600">{evt.eventId}</td>
                    <td className="px-4 py-3 text-sm font-medium text-gray-800">{evt.eventType}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
                      {format(new Date(evt.eventTime), 'HH:mm:ss.SSS')}
                    </td>
                    <td className="px-4 py-3">
                      <EventPayload attributes={evt.attributes} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {activeTab === 'input' && (
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-gray-500 text-sm">Input and result data display is not yet available for protobuf-encoded payloads.</p>
        </div>
      )}
    </div>
  );
}
