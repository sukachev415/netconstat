import { createContext, useContext, useState } from 'react';
import type { ReactNode } from 'react';

export type TimeInterval = '30m' | '1h' | '12h' | '1d' | '7d' | '30d';
export type DataUnit = 'KB' | 'MB' | 'GB';

interface DashboardContextType {
  interval: TimeInterval;
  setInterval: (interval: TimeInterval) => void;
  unit: DataUnit;
  setUnit: (unit: DataUnit) => void;
  sourceIp: string;
  setSourceIp: (ip: string) => void;
  service: string;
  setService: (service: string) => void;
  getTimeRange: () => { from: string; to: string };
}

const DashboardContext = createContext<DashboardContextType | undefined>(undefined);

function getIntervalDuration(interval: TimeInterval): number {
  switch (interval) {
    case '30m': return 30 * 60 * 1000;
    case '1h': return 60 * 60 * 1000;
    case '12h': return 12 * 60 * 60 * 1000;
    case '1d': return 24 * 60 * 60 * 1000;
    case '7d': return 7 * 24 * 60 * 60 * 1000;
    case '30d': return 30 * 24 * 60 * 60 * 1000;
  }
}

export function DashboardProvider({ children }: { children: ReactNode }) {
  const [interval, setInterval] = useState<TimeInterval>('1h');
  const [unit, setUnit] = useState<DataUnit>('MB');
  const [sourceIp, setSourceIp] = useState('');
  const [service, setService] = useState('');

  const getTimeRange = () => {
    const now = new Date();
    const from = new Date(now.getTime() - getIntervalDuration(interval));
    return {
      from: from.toISOString(),
      to: now.toISOString(),
    };
  };

  return (
    <DashboardContext.Provider
      value={{
        interval,
        setInterval,
        unit,
        setUnit,
        sourceIp,
        setSourceIp,
        service,
        setService,
        getTimeRange,
      }}
    >
      {children}
    </DashboardContext.Provider>
  );
}

export function useDashboard() {
  const context = useContext(DashboardContext);
  if (context === undefined) {
    throw new Error('useDashboard must be used within a DashboardProvider');
  }
  return context;
}
