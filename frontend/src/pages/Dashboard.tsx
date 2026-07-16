import { Header } from '../components/layout/Header';
import { TimeIntervalSwitch } from '../components/filters/TimeIntervalSwitch';
import { DeviceFilter } from '../components/filters/DeviceFilter';
import { ServiceFilter } from '../components/filters/ServiceFilter';
import { TrafficChart } from '../components/charts/TrafficChart';
import { TopServicesTable } from '../components/charts/TopServicesTable';
import { ProtocolBreakdown } from '../components/charts/ProtocolBreakdown';
import { ConnectionTimeline } from '../components/charts/ConnectionTimeline';

export function Dashboard() {
  return (
    <div className="min-h-screen bg-[var(--color-bg)]">
      <Header />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
          <div>
            <h2 className="text-2xl font-bold text-[var(--color-text)]">NetConStat</h2>
            <p className="text-sm text-[var(--color-text-secondary)] mt-1">
              NetFlow v9 Traffic Monitor
            </p>
          </div>
          <TimeIntervalSwitch />
        </div>

        <div className="flex flex-wrap gap-4 mb-6">
          <DeviceFilter />
          <ServiceFilter />
        </div>

        {/* Connection Timeline */}
        <div className="mb-6">
          <ConnectionTimeline />
        </div>

        {/* Charts row */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <TrafficChart />
          </div>
          <div>
            <ProtocolBreakdown />
          </div>
        </div>

        <div className="mt-6">
          <TopServicesTable />
        </div>
      </main>
    </div>
  );
}
