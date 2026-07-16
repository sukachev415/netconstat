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
}

const DashboardContext = createContext<DashboardContextType>({
  interval: '1h',
  setInterval: () => {},
  unit: 'MB',
  setUnit: () => {},
  sourceIp: '',
  setSourceIp: () => {},
  service: '',
  setService: () => {},
});

export function DashboardProvider({ children }: { children: ReactNode }) {
  const [interval, setInterval] = useState<TimeInterval>('1h');
  const [unit, setUnit] = useState<DataUnit>('MB');
  const [sourceIp, setSourceIp] = useState('');
  const [service, setService] = useState('');

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
      }}
    >
      {children}
    </DashboardContext.Provider>
  );
}

export function useDashboard() {
  return useContext(DashboardContext);
}
