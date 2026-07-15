import { useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { useTrafficStats } from '../../hooks/useStats';
import { useDashboard } from '../../contexts/DashboardContext';
import type { DataUnit } from '../../contexts/DashboardContext';
import { UnitToggle } from './UnitToggle';
import type { TrafficStat } from '../../api/client';

const COLORS = [
  '#0ea5e9', '#8b5cf6', '#f59e0b', '#10b981', '#ef4444',
  '#06b6d4', '#f97316', '#6366f1', '#14b8a6', '#e11d48',
];

function convertBytes(bytes: number, unit: DataUnit): number {
  switch (unit) {
    case 'KB': return bytes / 1024;
    case 'MB': return bytes / (1024 * 1024);
    case 'GB': return bytes / (1024 * 1024 * 1024);
  }
}

function formatTimestamp(ts: string): string {
  const date = new Date(ts);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function TrafficChart() {
  const { unit } = useDashboard();
  const { data: trafficStats, isLoading } = useTrafficStats();

  const { chartData, services } = useMemo(() => {
    if (!trafficStats) return { chartData: [], services: [] };

    const timeMap = new Map<string, Record<string, number | string>>();
    const serviceSet = new Set<string>();

    trafficStats.forEach((stat: TrafficStat) => {
      const time = formatTimestamp(stat.timestamp);
      serviceSet.add(stat.service);
      
      if (!timeMap.has(time)) {
        timeMap.set(time, { timestamp: stat.timestamp });
      }
      
      const entry = timeMap.get(time)!;
      entry[stat.service] = (Number(entry[stat.service]) || 0) + convertBytes(stat.bytes, unit);
    });

    const sortedServices = Array.from(serviceSet).slice(0, 8);
    const data = Array.from(timeMap.values()).sort(
      (a, b) => new Date(String(a.timestamp)).getTime() - new Date(String(b.timestamp)).getTime()
    );

    return { chartData: data, services: sortedServices };
  }, [trafficStats, unit]);

  if (isLoading) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[var(--color-text)]">Traffic Over Time</h3>
          <UnitToggle />
        </div>
        <div className="h-80 flex items-center justify-center text-[var(--color-text-secondary)]">
          Loading...
        </div>
      </div>
    );
  }

  return (
    <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-[var(--color-text)]">Traffic Over Time</h3>
        <UnitToggle />
      </div>
      <div className="h-80">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
            <XAxis
              dataKey="timestamp"
              tickFormatter={formatTimestamp}
              stroke="var(--color-text-secondary)"
              tick={{ fontSize: 12 }}
            />
            <YAxis
              stroke="var(--color-text-secondary)"
              tick={{ fontSize: 12 }}
              tickFormatter={(value: number) => `${value.toFixed(1)} ${unit}`}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: '8px',
                color: 'var(--color-text)',
              }}
              formatter={(value) => [`${Number(value ?? 0).toFixed(2)} ${unit}`]}
            />
            <Legend
              wrapperStyle={{ color: 'var(--color-text-secondary)' }}
            />
            {services.map((service, index) => (
              <Bar
                key={service}
                dataKey={service}
                stackId="a"
                fill={COLORS[index % COLORS.length]}
                radius={[2, 2, 0, 0]}
              />
            ))}
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
