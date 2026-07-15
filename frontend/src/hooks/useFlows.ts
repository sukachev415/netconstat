import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { Flow } from '../api/client';
import { useDashboard } from '../contexts/DashboardContext';

export function useFlows() {
  const { getTimeRange, sourceIp, service } = useDashboard();
  const { from, to } = getTimeRange();

  return useQuery({
    queryKey: ['flows', from, to, sourceIp, service],
    queryFn: () =>
      api.getFlows({
        from,
        to,
        source_ip: sourceIp || undefined,
        service: service || undefined,
        limit: 100,
      }),
    select: (data) => data.data as Flow[],
    refetchInterval: 30000,
  });
}
