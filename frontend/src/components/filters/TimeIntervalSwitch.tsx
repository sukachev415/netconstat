import { useDashboard } from '../../contexts/DashboardContext';
import type { TimeInterval } from '../../contexts/DashboardContext';

const intervals: { value: TimeInterval; label: string }[] = [
  { value: '30m', label: '30m' },
  { value: '1h', label: '1h' },
  { value: '12h', label: '12h' },
  { value: '1d', label: '1d' },
  { value: '7d', label: '7d' },
  { value: '30d', label: '30d' },
];

export function TimeIntervalSwitch() {
  const { interval, setInterval } = useDashboard();

  return (
    <div className="flex gap-1 bg-[var(--color-surface)] rounded-lg p-1 border border-[var(--color-border)]">
      {intervals.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => setInterval(value)}
          className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
            interval === value
              ? 'bg-brand-500 text-white'
              : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[var(--color-border)]'
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
