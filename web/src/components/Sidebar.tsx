import { Link, useParams, useNavigate } from 'react-router-dom';
import { useNamespaces } from '@/hooks/useNamespaces';
import { LayoutList, Zap } from 'lucide-react';

export function Sidebar() {
  const { namespace } = useParams<{ namespace: string }>();
  const navigate = useNavigate();
  const { data } = useNamespaces();
  const namespaces = data?.namespaces || [];
  const currentNs = namespace || namespaces[0]?.namespaceInfo.name || '';

  return (
    <aside className="w-60 bg-gray-900 text-gray-200 flex flex-col min-h-screen">
      <div className="p-4 border-b border-gray-700">
        <Link to="/" className="flex items-center gap-2 text-white font-bold text-lg">
          <Zap size={20} />
          Tempiex
        </Link>
      </div>

      <div className="p-3">
        <label className="text-xs text-gray-400 uppercase tracking-wide mb-1 block">Namespace</label>
        <select
          value={currentNs}
          onChange={(e) => navigate(`/namespaces/${e.target.value}/workflows`)}
          className="w-full bg-gray-800 border border-gray-600 rounded px-2 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          {namespaces.map((ns) => (
            <option key={ns.namespaceInfo.id} value={ns.namespaceInfo.name}>
              {ns.namespaceInfo.name}
            </option>
          ))}
        </select>
      </div>

      <nav className="flex-1 p-3 space-y-1">
        <Link
          to={`/namespaces/${currentNs}/workflows`}
          className="flex items-center gap-2 px-3 py-2 rounded text-sm hover:bg-gray-800 transition-colors"
        >
          <LayoutList size={16} />
          Workflows
        </Link>
        <span className="flex items-center gap-2 px-3 py-2 rounded text-sm text-gray-500 cursor-not-allowed">
          Schedules
        </span>
        <span className="flex items-center gap-2 px-3 py-2 rounded text-sm text-gray-500 cursor-not-allowed">
          Cluster
        </span>
        <Link
          to="/namespaces"
          className="flex items-center gap-2 px-3 py-2 rounded text-sm hover:bg-gray-800 transition-colors mt-4"
        >
          All Namespaces
        </Link>
      </nav>
    </aside>
  );
}
