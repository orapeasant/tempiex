import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Layout } from './components/Layout';
import { NamespaceList } from './pages/NamespaceList';
import { WorkflowList } from './pages/WorkflowList';
import { WorkflowDetail } from './pages/WorkflowDetail';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 5000,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Navigate to="/namespaces" replace />} />
            <Route path="/namespaces" element={<NamespaceList />} />
            <Route path="/namespaces/:namespace/workflows" element={<WorkflowList />} />
            <Route path="/namespaces/:namespace/workflows/:workflowId/:runId" element={<WorkflowDetail />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
