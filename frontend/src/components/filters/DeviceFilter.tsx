import { useDashboard } from '../../contexts/DashboardContext';
import { useDevices } from '../../hooks/useMetadata';

export function DeviceFilter() {
  const { sourceIp, setSourceIp } = useDashboard();
  const { data: devices } = useDevices();

  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-[var(--color-text-secondary)]">Source IP</label>
      <select
        value={sourceIp}
        onChange={(e) => setSourceIp(e.target.value)}
        className="px-3 py-2 text-sm bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
      >
        <option value="">All Devices</option>
        {devices?.map((device) => (
          <option key={device.ip} value={device.ip}>
            {device.hostname || device.ip}
          </option>
        ))}
      </select>
    </div>
  );
}
