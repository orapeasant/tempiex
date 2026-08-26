import { useEffect, useState } from 'react';
import { TempiexHttpClient, ClusterInfo } from '../client/http.js';

interface UseClusterResult {
  cluster: ClusterInfo | null;
  loading: boolean;
  error: Error | null;
}

export function useCluster(client: TempiexHttpClient): UseClusterResult {
  const [cluster, setCluster] = useState<ClusterInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client
      .getCluster()
      .then((res) => {
        if (!cancelled) {
          setCluster(res.clusterInfo);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [client]);

  return { cluster, loading, error };
}
