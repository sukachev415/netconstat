import { useTopServices } from '../../hooks/useStats';
import { useDashboard } from '../../contexts/DashboardContext';
import type { DataUnit } from '../../contexts/DashboardContext';
import type { TopService } from '../../api/client';

function formatBytes(bytes: number, unit: DataUnit): string {
  let value: number;
  switch (unit) {
    case 'KB': value = bytes / 1024; break;
    case 'MB': value = bytes / (1024 * 1024); break;
    case 'GB': value = bytes / (1024 * 1024 * 1024); break;
  }
  return `${value.toFixed(2)} ${unit}`;
}

const COLORS = [
  '#0ea5e9', '#8b5cf6', '#f59e0b', '#10b981', '#ef4444',
  '#06b6d4', '#f97316', '#6366f1', '#14b8a6', '#e11d48',
];

export function TopServicesTable() {
  const { unit } = useDashboard();
  const { data: topServices, isLoading } = useTopServices();

  const totalBytes = topServices?.reduce((sum: number, s: TopService) => sum + s.total_bytes, 0) || 1;

  if (isLoading) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Top Services</h3>
        <div className="h-60 flex items-center justify-center text-[var(--color-text-secondary)]">
          Loading...
        </div>
      </div>
    );
  }

  if (!topServices || topServices.length === 0) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Top Services</h3>
        <div className="h-60 flex items-center justify-center text-[var(--color-text-secondary)]">
          No data yet...
        </div>
      </div>
    );
  }

  return (
    <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
      <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Top Services</h3>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-[var(--color-border)]">
              <th className="text-left py-3 px-4 text-sm font-medium text-[var(--color-text-secondary)]">Service</th>
              <th className="text-right py-3 px-4 text-sm font-medium text-[var(--color-text-secondary)]">Traffic</th>
              <th className="text-right py-3 px-4 text-sm font-medium text-[var(--color-text-secondary)]">Flows</th>
              <th className="text-right py-3 px-4 text-sm font-medium text-[var(--color-text-secondary)]">Share</th>
            </tr>
          </thead>
          <tbody>
            {topServices.map((service: TopService, index: number) => {
              const pct = (service.total_bytes / totalBytes * 100);
              return (
                <tr
                  key={service.service}
                  className="border-b border-[var(--color-border)] last:border-0 hover:bg-[var(--color-border)] transition-colors"
                >
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded-full" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                      <span className="text-sm font-medium text-[var(--color-text)]">{service.service}</span>
                    </div>
                  </td>
                  <td className="text-right py-3 px-4 text-sm text-[var(--color-text)]">
                    {formatBytes(service.total_bytes, unit)}
                  </td>
                  <td className="text-right py-3 px-4 text-sm text-[var(--color-text)]">
                    {service.flow_count.toLocaleString()}
                  </td>
                  <td className="text-right py-3 px-4">
                    <div className="flex items-center justify-end gap-2">
                      <div className="w-20 h-2 bg-[var(--color-border)] rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full"
                          style={{
                            width: `${pct}%`,
                            backgroundColor: COLORS[index % COLORS.length],
                          }}
                        />
                      </div>
                      <span className="text-sm text-[var(--color-text-secondary)]">
                        {pct.toFixed(1)}%
                      </span>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
