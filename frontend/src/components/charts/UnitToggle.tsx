import { useDashboard } from '../../contexts/DashboardContext';
import type { DataUnit } from '../../contexts/DashboardContext';

const units: DataUnit[] = ['KB', 'MB', 'GB'];

export function UnitToggle() {
  const { unit, setUnit } = useDashboard();

  return (
    <div className="flex gap-1 bg-[var(--color-surface)] rounded-lg p-1 border border-[var(--color-border)]">
      {units.map((u) => (
        <button
          key={u}
          onClick={() => setUnit(u)}
          className={`px-2 py-1 text-xs font-medium rounded-md transition-colors ${
            unit === u
              ? 'bg-brand-500 text-white'
              : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[var(--color-border)]'
          }`}
        >
          {u}
        </button>
      ))}
    </div>
  );
}
