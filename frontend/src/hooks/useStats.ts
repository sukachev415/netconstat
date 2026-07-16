import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { TrafficStat, TopService, ProtocolStat } from '../api/client';
import { useDashboard } from '../contexts/DashboardContext';

export function useTrafficStats() {
  const { interval } = useDashboard();

  return useQuery({
    queryKey: ['traffic-stats', interval],
    queryFn: async () => {
      const now = new Date();
      const durations: Record<string, number> = {
        '30m': 30*60*1000, '1h': 60*60*1000, '12h': 12*60*60*1000,
        '1d': 24*60*60*1000, '7d': 7*24*60*60*1000, '30d': 30*24*60*60*1000,
      };
      const intervalMap: Record<string, string> = {
        '30m': '5m', '1h': '5m', '12h': '30m', '1d': '1h', '7d': '6h', '30d': '1d',
      };
      const from = new Date(now.getTime() - (durations[interval] || 3600000)).toISOString();
      const to = now.toISOString();
      const res = await api.getTrafficStats({
        from, to,
        interval: intervalMap[interval] || '5m',
        group_by: 'service',
      });
      return (res.data || []) as TrafficStat[];
    },
    refetchInterval: 30000,
  });
}

export function useTopServices() {
  const { interval } = useDashboard();

  return useQuery({
    queryKey: ['top-services', interval],
    queryFn: async () => {
      const now = new Date();
      const durations: Record<string, number> = {
        '30m': 30*60*1000, '1h': 60*60*1000, '12h': 12*60*60*1000,
        '1d': 24*60*60*1000, '7d': 7*24*60*60*1000, '30d': 30*24*60*60*1000,
      };
      const from = new Date(now.getTime() - (durations[interval] || 3600000)).toISOString();
      const to = now.toISOString();
      const res = await api.getTopServices({ from, to, top: 10 });
      return (res.data || []) as TopService[];
    },
    refetchInterval: 30000,
  });
}

export function useProtocols() {
  const { interval } = useDashboard();

  return useQuery({
    queryKey: ['protocols', interval],
    queryFn: async () => {
      const now = new Date();
      const durations: Record<string, number> = {
        '30m': 30*60*1000, '1h': 60*60*1000, '12h': 12*60*60*1000,
        '1d': 24*60*60*1000, '7d': 7*24*60*60*1000, '30d': 30*24*60*60*1000,
      };
      const from = new Date(now.getTime() - (durations[interval] || 3600000)).toISOString();
      const to = now.toISOString();
      const res = await api.getProtocols({ from, to });
      return (res.data || []) as ProtocolStat[];
    },
    refetchInterval: 30000,
  });
}
