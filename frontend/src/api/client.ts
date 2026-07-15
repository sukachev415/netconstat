const API_BASE = '/api/v1';

interface ApiResponse<T> {
  data: T;
  meta?: Record<string, unknown>;
}

export async function fetchApi<T>(
  endpoint: string,
  params?: Record<string, string | number | undefined>
): Promise<ApiResponse<T>> {
  const url = new URL(`${API_BASE}${endpoint}`, window.location.origin);
  
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        url.searchParams.append(key, String(value));
      }
    });
  }

  const response = await fetch(url.toString());
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

// Types
export interface Flow {
  id: number;
  timestamp: string;
  source_ip: string;
  dest_ip: string;
  source_port: number;
  dest_port: number;
  protocol: string;
  service: string;
  bytes_sent: number;
  bytes_received: number;
  packets_sent: number;
  packets_received: number;
  duration: number;
}

export interface TrafficStat {
  timestamp: string;
  service: string;
  bytes: number;
  packets: number;
  flows: number;
}

export interface TopService {
  service: string;
  total_bytes: number;
  total_packets: number;
  total_flows: number;
  percentage: number;
}

export interface ProtocolStat {
  protocol: string;
  total_bytes: number;
  total_packets: number;
  total_flows: number;
  percentage: number;
}

export interface Service {
  name: string;
  port: number;
  protocol: string;
}

export interface Device {
  ip: string;
  hostname?: string;
  last_seen: string;
}

// API functions
export const api = {
  getFlows: (params?: {
    from?: string;
    to?: string;
    source_ip?: string;
    dest_ip?: string;
    service?: string;
    protocol?: string;
    limit?: number;
    offset?: number;
  }) => fetchApi<Flow[]>('/flows', params as Record<string, string | number | undefined>),

  getTrafficStats: (params?: {
    from?: string;
    to?: string;
    interval?: string;
    group_by?: string;
  }) => fetchApi<TrafficStat[]>('/stats/traffic', params as Record<string, string | number | undefined>),

  getTopServices: (params?: {
    from?: string;
    to?: string;
    top?: number;
  }) => fetchApi<TopService[]>('/stats/top-services', params as Record<string, string | number | undefined>),

  getProtocols: (params?: {
    from?: string;
    to?: string;
  }) => fetchApi<ProtocolStat[]>('/stats/protocols', params as Record<string, string | number | undefined>),

  getServices: () => fetchApi<Service[]>('/services'),
  
  getDevices: () => fetchApi<Device[]>('/devices'),
};
