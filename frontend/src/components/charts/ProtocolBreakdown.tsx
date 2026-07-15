import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { useProtocols } from '../../hooks/useStats';
import { useDashboard } from '../../contexts/DashboardContext';
import type { DataUnit } from '../../contexts/DashboardContext';

const COLORS = [
  '#0ea5e9', '#8b5cf6', '#f59e0b', '#10b981', '#ef4444',
  '#06b6d4', '#f97316', '#6366f1', '#14b8a6', '#e11d48',
];

function formatBytes(bytes: number, unit: DataUnit): string {
  let value: number;
  switch (unit) {
    case 'KB': value = bytes / 1024; break;
    case 'MB': value = bytes / (1024 * 1024); break;
    case 'GB': value = bytes / (1024 * 1024 * 1024); break;
  }
  return `${value.toFixed(2)} ${unit}`;
}

interface ProtocolData {
  name: string;
  value: number;
  percentage: number;
}

export function ProtocolBreakdown() {
  const { unit } = useDashboard();
  const { data: protocols, isLoading } = useProtocols();

  const chartData: ProtocolData[] = protocols?.map((p) => ({
    name: p.protocol,
    value: p.total_bytes,
    percentage: p.percentage,
  })) || [];

  if (isLoading) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Protocol Distribution</h3>
        <div className="h-60 flex items-center justify-center text-[var(--color-text-secondary)]">
          Loading...
        </div>
      </div>
    );
  }

  return (
    <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
      <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Protocol Distribution</h3>
      <div className="h-60">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={80}
              paddingAngle={2}
              dataKey="value"
            >
              {chartData.map((_, index) => (
                <Cell key={index} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: '8px',
                color: 'var(--color-text)',
              }}
              formatter={(value) => [formatBytes(Number(value ?? 0), unit)]}
            />
            <Legend
              wrapperStyle={{ color: 'var(--color-text-secondary)' }}
              formatter={(value: string) => <span style={{ color: 'var(--color-text-secondary)' }}>{value}</span>}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
      <div className="mt-4 grid grid-cols-2 gap-2">
        {chartData.slice(0, 4).map((item, index) => (
          <div key={item.name} className="flex items-center gap-2">
            <span
              className="w-2 h-2 rounded-full"
              style={{ backgroundColor: COLORS[index % COLORS.length] }}
            />
            <span className="text-xs text-[var(--color-text-secondary)] truncate">
              {item.name}: {item.percentage.toFixed(1)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
