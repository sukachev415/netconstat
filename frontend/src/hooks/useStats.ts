import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { TrafficStat, TopService, ProtocolStat } from '../api/client';
import { useDashboard } from '../contexts/DashboardContext';

export function useTrafficStats() {
  const { getTimeRange, interval, sourceIp } = useDashboard();
  const { from, to } = getTimeRange();

  return useQuery({
    queryKey: ['traffic-stats', from, to, interval, sourceIp],
    queryFn: () =>
      api.getTrafficStats({
        from,
        to,
        interval: interval === '30m' ? '5m' : interval === '1h' ? '5m' : interval === '12h' ? '30m' : interval === '1d' ? '1h' : interval === '7d' ? '6h' : '1d',
        group_by: 'service',
      }),
    select: (data) => data.data as TrafficStat[],
    refetchInterval: 30000,
  });
}

export function useTopServices() {
  const { getTimeRange, sourceIp } = useDashboard();
  const { from, to } = getTimeRange();

  return useQuery({
    queryKey: ['top-services', from, to, sourceIp],
    queryFn: () =>
      api.getTopServices({
        from,
        to,
        top: 10,
      }),
    select: (data) => data.data as TopService[],
    refetchInterval: 30000,
  });
}

export function useProtocols() {
  const { getTimeRange, sourceIp } = useDashboard();
  const { from, to } = getTimeRange();

  return useQuery({
    queryKey: ['protocols', from, to, sourceIp],
    queryFn: () =>
      api.getProtocols({
        from,
        to,
      }),
    select: (data) => data.data as ProtocolStat[],
    refetchInterval: 30000,
  });
}
