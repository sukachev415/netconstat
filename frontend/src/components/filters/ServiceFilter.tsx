import { useDashboard } from '../../contexts/DashboardContext';
import { useServices } from '../../hooks/useMetadata';

export function ServiceFilter() {
  const { service, setService } = useDashboard();
  const { data: services } = useServices();

  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-[var(--color-text-secondary)]">Service</label>
      <select
        value={service}
        onChange={(e) => setService(e.target.value)}
        className="px-3 py-2 text-sm bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
      >
        <option value="">All Services</option>
        {services?.map((svc) => (
          <option key={svc} value={svc}>{svc}</option>
        ))}
      </select>
    </div>
  );
}
