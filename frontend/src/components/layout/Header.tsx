import { ThemeToggle } from '../ui/ThemeToggle';

export function Header() {
  return (
    <header className="bg-[var(--color-surface)] border-b border-[var(--color-border)] px-6 py-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-brand-500 rounded-lg flex items-center justify-center">
            <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--color-text)]">NetConstat</h1>
            <p className="text-xs text-[var(--color-text-secondary)]">Network Traffic Dashboard</p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="hidden sm:flex items-center gap-2 text-sm text-[var(--color-text-secondary)]">
            <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
            Live
          </div>
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
