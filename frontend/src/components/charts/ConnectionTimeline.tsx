import { useState } from 'react';
import { useFlows } from '../../hooks/useFlows';
import type { Flow } from '../../api/client';

function formatTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function protocolName(p: number): string {
  switch (p) {
    case 1: return 'ICMP';
    case 6: return 'TCP';
    case 17: return 'UDP';
    case 47: return 'GRE';
    case 50: return 'ESP';
    default: return `${p}`;
  }
}

export function ConnectionTimeline() {
  const { data: flows, isLoading } = useFlows();
  const [copied, setCopied] = useState<string | null>(null);

  const copyIP = (ip: string) => {
    navigator.clipboard.writeText(ip);
    setCopied(ip);
    setTimeout(() => setCopied(null), 1500);
  };

  if (isLoading) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Connection Timeline</h3>
        <div className="h-40 flex items-center justify-center text-[var(--color-text-secondary)]">Loading...</div>
      </div>
    );
  }

  if (!flows || flows.length === 0) {
    return (
      <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
        <h3 className="text-lg font-semibold text-[var(--color-text)] mb-4">Connection Timeline</h3>
        <div className="h-40 flex items-center justify-center text-[var(--color-text-secondary)]">No connections yet...</div>
      </div>
    );
  }

  // Group by dest_ip to show unique destinations
  const destMap = new Map<string, { service: string; bytes: number; flows: Flow[]; lastSeen: string; destPort: number; protocol: number }>();
  flows.forEach((f: Flow) => {
    const key = f.dest_ip;
    const existing = destMap.get(key);
    if (existing) {
      existing.bytes += f.bytes;
      existing.flows.push(f);
      if (f.timestamp > existing.lastSeen) {
        existing.lastSeen = f.timestamp;
        existing.destPort = f.dest_port;
        existing.protocol = f.protocol;
      }
    } else {
      destMap.set(key, {
        service: f.service || 'Unknown',
        bytes: f.bytes,
        flows: [f],
        lastSeen: f.timestamp,
        destPort: f.dest_port,
        protocol: f.protocol,
      });
    }
  });

  // Sort by last seen, most recent first
  const entries = Array.from(destMap.entries())
    .sort((a, b) => b[1].lastSeen.localeCompare(a[1].lastSeen));

  return (
    <div className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-[var(--color-text)]">
          Connection Timeline
          <span className="ml-2 text-sm font-normal text-[var(--color-text-secondary)]">
            {entries.length} destinations
          </span>
        </h3>
      </div>

      <div className="space-y-1 max-h-[600px] overflow-y-auto">
        {entries.map(([destIp, info]) => (
          <div
            key={destIp}
            className="flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-[var(--color-border)] transition-colors group"
          >
            {/* Time */}
            <span className="text-xs text-[var(--color-text-secondary)] w-20 shrink-0 font-mono">
              {formatTime(info.lastSeen)}
            </span>

            {/* Timeline dot */}
            <span className="w-2 h-2 rounded-full bg-sky-500 shrink-0" />

            {/* Service name */}
            <span className="text-sm font-medium text-[var(--color-text)] w-48 shrink-0 truncate" title={info.service}>
              {info.service}
            </span>

            {/* Dest IP - clickable to copy */}
            <button
              onClick={() => copyIP(destIp)}
              className="text-sm font-mono text-sky-400 hover:text-sky-300 w-36 shrink-0 text-left transition-colors"
              title="Click to copy"
            >
              {copied === destIp ? (
                <span className="text-green-400">Copied!</span>
              ) : (
                destIp
              )}
            </button>

            {/* Port + Protocol */}
            <span className="text-xs text-[var(--color-text-secondary)] w-24 shrink-0 font-mono">
              :{info.destPort} {protocolName(info.protocol)}
            </span>

            {/* Traffic */}
            <span className="text-xs text-[var(--color-text-secondary)] w-20 shrink-0 text-right">
              {formatBytes(info.bytes)}
            </span>

            {/* Flow count */}
            <span className="text-xs text-[var(--color-text-secondary)] w-16 shrink-0 text-right">
              {info.flows.length} flows
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
