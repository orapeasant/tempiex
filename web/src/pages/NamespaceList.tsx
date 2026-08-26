import { useNavigate } from 'react-router-dom';
import { useNamespaces } from '@/hooks/useNamespaces';

export function NamespaceList() {
  const { data, isLoading } = useNamespaces();
  const navigate = useNavigate();

  if (isLoading) return <div className="text-gray-500">Loading namespaces...</div>;

  const namespaces = data?.namespaces || [];

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Namespaces</h1>
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-800 text-white">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Name</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Retention (days)</th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase">Description</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {namespaces.map((ns) => (
              <tr
                key={ns.namespaceInfo.id}
                onClick={() => navigate(`/namespaces/${ns.namespaceInfo.name}/workflows`)}
                className="hover:bg-blue-50 cursor-pointer transition-colors"
              >
                <td className="px-4 py-3 text-sm font-medium text-blue-600">{ns.namespaceInfo.name}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{ns.config.workflowExecutionRetentionTtl.days}</td>
                <td className="px-4 py-3 text-sm text-gray-500">{ns.namespaceInfo.description || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
