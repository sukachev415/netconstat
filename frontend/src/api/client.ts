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

// Types matching actual API responses
export interface Flow {
  timestamp: string;
  source_ip: string;
  dest_ip: string;
  source_port: number;
  dest_port: number;
  protocol: number;
  traffic_mark: number;
  bytes: number;
  packets: number;
  service: string;
}

export interface TrafficStat {
  bucket: string;
  label: string;
  total_bytes: number;
  total_packets: number;
  flow_count: number;
}

export interface TopService {
  service: string;
  total_bytes: number;
  total_packets: number;
  flow_count: number;
}

export interface ProtocolStat {
  protocol: number;
  name: string;
  total_bytes: number;
  flow_count: number;
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

  getServices: () => fetchApi<string[]>('/services'),
  
  getDevices: () => fetchApi<string[]>('/devices'),
};
