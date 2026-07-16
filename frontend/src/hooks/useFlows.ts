import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { Flow } from '../api/client';
import { useDashboard } from '../contexts/DashboardContext';

export function useFlows() {
  const { interval, sourceIp, service } = useDashboard();

  return useQuery({
    queryKey: ['flows', interval, sourceIp, service],
    queryFn: async () => {
      const now = new Date();
      const durations: Record<string, number> = {
        '30m': 30*60*1000, '1h': 60*60*1000, '12h': 12*60*60*1000,
        '1d': 24*60*60*1000, '7d': 7*24*60*60*1000, '30d': 30*24*60*60*1000,
      };
      const from = new Date(now.getTime() - (durations[interval] || 3600000)).toISOString();
      const to = now.toISOString();
      const res = await api.getFlows({
        from,
        to,
        source_ip: sourceIp || undefined,
        service: service || undefined,
        limit: 100,
      });
      return (res.data || []) as Flow[];
    },
    refetchInterval: 30000,
  });
}
