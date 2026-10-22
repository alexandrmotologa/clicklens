import React, { useState, useEffect } from 'react';
import { 
  Search, 
  Layers, 
  Zap, 
  Terminal, 
  Cpu, 
  Clock, 
  HardDrive, 
  FileText, 
  Database,
  ArrowRight,
  Sparkles,
  Command
} from 'lucide-react';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab: (tab: string) => void;
  onOpenReport: () => void;
  onToggleDemo: () => void;
  onSelectTable: (db: string, tbl: string) => void;
}

interface CommandItem {
  id: string;
  category: 'Navigation' | 'Actions' | 'Databases';
  title: string;
  subtitle: string;
  icon: React.ReactNode;
  action: () => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
  onOpenReport,
  onToggleDemo,
  onSelectTable,
}) => {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  const commands: CommandItem[] = [
    {
      id: 'nav-parts',
      category: 'Navigation',
      title: 'MergeTree Parts & Heatmap',
      subtitle: 'Inspect partition fragmentation, part age histogram, and active chunks',
      icon: <Layers className="w-4 h-4 text-sky-400" />,
      action: () => { onNavigateTab('parts'); onClose(); },
    },
    {
      id: 'nav-merges',
      category: 'Navigation',
      title: 'Background Merges & Mutations',
      subtitle: 'Real-time telemetry on system.merges and asynchronous table mutations',
      icon: <Zap className="w-4 h-4 text-amber-400" />,
      action: () => { onNavigateTab('merges'); onClose(); },
    },
    {
      id: 'nav-codecs',
      category: 'Navigation',
      title: 'Column Codec Optimizer',
      subtitle: 'Analyze compression efficiency (ZSTD, Gorilla, DoubleDelta) and playbooks',
      icon: <HardDrive className="w-4 h-4 text-emerald-400" />,
      action: () => { onNavigateTab('codecs'); onClose(); },
    },
    {
      id: 'nav-profiler',
      category: 'Navigation',
      title: 'Query Pipeline Profiler',
      subtitle: 'Visual EXPLAIN PIPELINE DAG, thread concurrency, and disk spill alerts',
      icon: <Cpu className="w-4 h-4 text-indigo-400" />,
      action: () => { onNavigateTab('profiler'); onClose(); },
    },
    {
      id: 'nav-query-log',
      category: 'Navigation',
      title: 'Historical Query Log Analytics',
      subtitle: 'Filter slowest queries, peak RAM spenders, disk spills, and exceptions',
      icon: <Clock className="w-4 h-4 text-purple-400" />,
      action: () => { onNavigateTab('query-log'); onClose(); },
    },
    {
      id: 'nav-workbench',
      category: 'Navigation',
      title: 'Serverless Local File Workbench',
      subtitle: 'Inspect and query local Parquet, CSV, and TSV files via clickhouse-local',
      icon: <Terminal className="w-4 h-4 text-cyan-400" />,
      action: () => { onNavigateTab('workbench'); onClose(); },
    },
    {
      id: 'nav-console',
      category: 'Navigation',
      title: 'SQL Query Console',
      subtitle: 'Interactive query editor with tabular execution matrix',
      icon: <Database className="w-4 h-4 text-slate-300" />,
      action: () => { onNavigateTab('console'); onClose(); },
    },
    {
      id: 'act-report',
      category: 'Actions',
      title: 'Generate Health Audit Report',
      subtitle: 'Export comprehensive diagnostic audit as formatted Markdown or JSON',
      icon: <FileText className="w-4 h-4 text-rose-400" />,
      action: () => { onOpenReport(); onClose(); },
    },
    {
      id: 'act-demo',
      category: 'Actions',
      title: 'Toggle Demo / Live Mode',
      subtitle: 'Switch between simulated cluster telemetry and active ClickHouse server',
      icon: <Sparkles className="w-4 h-4 text-amber-400" />,
      action: () => { onToggleDemo(); onClose(); },
    },
    {
      id: 'db-telemetry',
      category: 'Databases',
      title: 'Select Table: telemetry.events',
      subtitle: 'High-throughput event stream with multi-level MergeTree partitions',
      icon: <Database className="w-4 h-4 text-sky-400" />,
      action: () => { onSelectTable('telemetry', 'events'); onClose(); },
    },
    {
      id: 'db-analytics',
      category: 'Databases',
      title: 'Select Table: analytics.orders',
      subtitle: 'E-commerce transactional partition dataset',
      icon: <Database className="w-4 h-4 text-sky-400" />,
      action: () => { onSelectTable('analytics', 'orders'); onClose(); },
    },
  ];

  const filteredCommands = commands.filter((c) => {
    if (!query) return true;
    const lower = query.toLowerCase();
    return c.title.toLowerCase().includes(lower) || c.subtitle.toLowerCase().includes(lower);
  });

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev + 1) % (filteredCommands.length || 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev - 1 + filteredCommands.length) % (filteredCommands.length || 1));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].action();
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, filteredCommands, selectedIndex, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-start justify-center pt-24 p-4">
      <div className="bg-[#10131a] border border-[#232938] rounded-2xl max-w-xl w-full shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-100">
        {/* Search Input Bar */}
        <div className="p-4 border-b border-[#1e2330] flex items-center gap-3">
          <Search className="w-5 h-5 text-slate-400 shrink-0" />
          <input
            autoFocus
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Type a command or jump to feature (e.g. Parts, Codecs, Query Log, Report)..."
            className="w-full bg-transparent text-sm text-white placeholder-slate-500 focus:outline-none"
          />
          <kbd className="hidden sm:inline-block px-2 py-0.5 bg-[#181c26] border border-[#283144] rounded text-[10px] font-mono text-slate-400">
            ESC
          </kbd>
        </div>

        {/* Results List */}
        <div className="max-h-80 overflow-y-auto p-2 space-y-1">
          {filteredCommands.length === 0 ? (
            <div className="p-6 text-center text-xs text-slate-500">
              No matching commands found.
            </div>
          ) : (
            filteredCommands.map((cmd, idx) => {
              const isSelected = idx === selectedIndex;
              return (
                <div
                  key={cmd.id}
                  onClick={() => cmd.action()}
                  onMouseEnter={() => setSelectedIndex(idx)}
                  className={`p-3 rounded-xl flex items-center justify-between cursor-pointer transition-colors ${
                    isSelected ? 'bg-indigo-600/20 text-white' : 'hover:bg-[#161a24] text-slate-300'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-[#181c26] border border-[#252c3d]">
                      {cmd.icon}
                    </div>
                    <div>
                      <div className="text-xs font-semibold text-white flex items-center gap-2">
                        <span>{cmd.title}</span>
                        <span className="text-[10px] font-normal px-1.5 py-0.2 rounded bg-[#1e2330] text-slate-400">
                          {cmd.category}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-400 line-clamp-1">{cmd.subtitle}</p>
                    </div>
                  </div>

                  <ArrowRight className={`w-3.5 h-3.5 opacity-60 ${isSelected ? 'text-indigo-400' : 'text-slate-600'}`} />
                </div>
              );
            })
          )}
        </div>

        {/* Footer shortcuts */}
        <div className="p-2.5 bg-[#0b0c10] border-t border-[#181c24] flex items-center justify-between text-[11px] text-slate-500 font-mono px-4">
          <div className="flex items-center gap-3">
            <span>↑↓ Navigate</span>
            <span>↵ Select</span>
            <span>ESC Close</span>
          </div>
          <div className="flex items-center gap-1 text-slate-400">
            <Command className="w-3 h-3" />
            <span>ClickLens Quick Jump</span>
          </div>
        </div>
      </div>
    </div>
  );
};
