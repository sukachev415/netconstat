import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useServices() {
  return useQuery({
    queryKey: ['services'],
    queryFn: () => api.getServices(),
    select: (data) => data.data as string[],
    staleTime: 5 * 60 * 1000,
  });
}

export function useDevices() {
  return useQuery({
    queryKey: ['devices'],
    queryFn: () => api.getDevices(),
    select: (data) => data.data as string[],
    staleTime: 5 * 60 * 1000,
  });
}
