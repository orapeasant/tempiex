import { useEffect, useState } from 'react';
import { TempiexHttpClient, Namespace } from '../client/http.js';

interface UseNamespacesResult {
  namespaces: Namespace[];
  loading: boolean;
  error: Error | null;
}

export function useNamespaces(client: TempiexHttpClient): UseNamespacesResult {
  const [namespaces, setNamespaces] = useState<Namespace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    client
      .listNamespaces()
      .then((res) => {
        if (!cancelled) {
          setNamespaces(res.namespaces ?? []);
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

  return { namespaces, loading, error };
}
