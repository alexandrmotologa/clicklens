import { useState, useEffect, useCallback } from 'react';
import { useClickHouseStream } from './hooks/useClickHouseStream';
import type { DatabaseInfo, TablePartsResponse, PipelineAnalysis, QueryResult } from './types';
import { Header } from './components/Header';
import { PartsHeatmap } from './components/PartsHeatmap';
import { MergesMonitor } from './components/MergesMonitor';
import { PipelineGraph } from './components/PipelineGraph';
import { CodecOptimizer } from './components/CodecOptimizer';
import { LocalWorkbench } from './components/LocalWorkbench';
import { QueryConsole } from './components/QueryConsole';
import { QueryLogViewer } from './components/QueryLogViewer';
import { CommandPalette } from './components/CommandPalette';
import { Cpu, Database, Flame, HardDrive, Layers, Terminal, Clock } from 'lucide-react';

export function App() {
  const { merges, status, refetchStatus, refetchMerges } = useClickHouseStream();

  const [activeTab, setActiveTab] = useState<'PARTS' | 'MERGES' | 'PIPELINE' | 'CODECS' | 'LOGS' | 'LOCAL' | 'CONSOLE'>('PARTS');
  const [databases, setDatabases] = useState<DatabaseInfo[]>([]);
  const [selectedDb, setSelectedDb] = useState<string>('telemetry');
  const [selectedTable, setSelectedTable] = useState<string>('events');

  const [partsData, setPartsData] = useState<TablePartsResponse | null>(null);
  const [loadingParts, setLoadingParts] = useState<boolean>(true);

  // Command palette state
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);

  // Load databases
  const fetchDatabases = useCallback(async () => {
    try {
      const res = await fetch('/api/databases');
      if (res.ok) {
        const data: DatabaseInfo[] = await res.json();
        setDatabases(data);
        if (data.length > 0 && !data.find((d) => d.name === selectedDb)) {
          setSelectedDb(data[0].name);
          if (data[0].tables && data[0].tables.length > 0) {
            setSelectedTable(data[0].tables[0]);
          }
        }
      }
    } catch {
      // Ignored
    }
  }, [selectedDb]);

  // Load parts for active table
  const fetchParts = useCallback(async (db: string, tbl: string) => {
    try {
      setLoadingParts(true);
      const res = await fetch(`/api/parts?database=${db}&table=${tbl}`);
      if (res.ok) {
        const data: TablePartsResponse = await res.json();
        setPartsData(data);
      }
    } catch {
      // Ignored
    } finally {
      setLoadingParts(false);
    }
  }, []);

  useEffect(() => {
    fetchDatabases();
  }, [fetchDatabases]);

  useEffect(() => {
    if (selectedDb && selectedTable) {
      fetchParts(selectedDb, selectedTable);
    }
  }, [selectedDb, selectedTable, fetchParts]);

  // Listen for Ctrl+K / Cmd+K global shortcut
  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setIsCommandPaletteOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleGlobalKeyDown);
    return () => window.removeEventListener('keydown', handleGlobalKeyDown);
  }, []);

  const handleConnect = async (dsn: string) => {
    try {
      await fetch('/api/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dsn }),
      });
      refetchStatus();
      refetchMerges();
      fetchDatabases();
    } catch {
      // Ignored
    }
  };

  const handleSelectDb = (db: string) => {
    setSelectedDb(db);
    const dbObj = databases.find((d) => d.name === db);
    if (dbObj && dbObj.tables && dbObj.tables.length > 0) {
      setSelectedTable(dbObj.tables[0]);
    }
  };

  const handleProfileQuery = async (query: string, mode: string): Promise<PipelineAnalysis | null> => {
    try {
      const res = await fetch('/api/explain', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, mode }),
      });
      if (res.ok) {
        return await res.json();
      }
    } catch {
      // Ignored
    }
    return null;
  };

  const handleExecuteQuery = async (query: string): Promise<QueryResult | null> => {
    try {
      const res = await fetch('/api/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
      });
      return await res.json();
    } catch (e: any) {
      return {
        query,
        columns: [],
        rows: [],
        rows_count: 0,
        execution_time_ms: 0,
        error: e.message || 'Execution error',
      };
    }
  };

  const handleNavigateFromCommand = (tabName: string) => {
    switch (tabName) {
      case 'parts': setActiveTab('PARTS'); break;
      case 'merges': setActiveTab('MERGES'); break;
      case 'codecs': setActiveTab('CODECS'); break;
      case 'profiler': setActiveTab('PIPELINE'); break;
      case 'query-log': setActiveTab('LOGS'); break;
      case 'workbench': setActiveTab('LOCAL'); break;
      case 'console': setActiveTab('CONSOLE'); break;
    }
  };

  return (
    <div className="min-h-screen bg-[#090d16] text-slate-100 flex flex-col font-sans selection:bg-sky-500/30 selection:text-sky-200">
      <Header
        status={status}
        databases={databases}
        selectedDb={selectedDb}
        selectedTable={selectedTable}
        onSelectDb={handleSelectDb}
        onSelectTable={setSelectedTable}
        onConnect={handleConnect}
        onRefresh={() => {
          refetchStatus();
          refetchMerges();
          fetchParts(selectedDb, selectedTable);
        }}
        onOpenCommandPalette={() => setIsCommandPaletteOpen(true)}
      />

      {/* Main Tab Navigation */}
      <div className="border-b border-slate-800 bg-[#0d1322] px-6">
        <div className="max-w-7xl mx-auto flex items-center gap-1.5 overflow-x-auto py-2 scrollbar-none">
          <button
            onClick={() => setActiveTab('PARTS')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'PARTS'
                ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40 shadow-sm shadow-sky-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Layers className="w-3.5 h-3.5" />
            <span>Parts & Partitions</span>
          </button>

          <button
            onClick={() => setActiveTab('MERGES')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'MERGES'
                ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40 shadow-sm shadow-amber-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Flame className="w-3.5 h-3.5" />
            <span>Background Merges</span>
            {merges.length > 0 && (
              <span className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
            )}
          </button>

          <button
            onClick={() => setActiveTab('PIPELINE')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'PIPELINE'
                ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40 shadow-sm shadow-sky-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Cpu className="w-3.5 h-3.5" />
            <span>Pipeline DAG</span>
          </button>

          <button
            onClick={() => setActiveTab('LOGS')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'LOGS'
                ? 'bg-purple-500/20 text-purple-300 border border-purple-500/40 shadow-sm shadow-purple-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Clock className="w-3.5 h-3.5" />
            <span>Historical Query Log</span>
          </button>

          <button
            onClick={() => setActiveTab('CODECS')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'CODECS'
                ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 shadow-sm shadow-emerald-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <HardDrive className="w-3.5 h-3.5" />
            <span>Codec Optimizer</span>
          </button>

          <button
            onClick={() => setActiveTab('LOCAL')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'LOCAL'
                ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40 shadow-sm shadow-amber-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Database className="w-3.5 h-3.5" />
            <span>Local Workbench</span>
          </button>

          <button
            onClick={() => setActiveTab('CONSOLE')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all shrink-0 cursor-pointer ${
              activeTab === 'CONSOLE'
                ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40 shadow-sm shadow-sky-500/10'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Terminal className="w-3.5 h-3.5" />
            <span>SQL Console</span>
          </button>
        </div>
      </div>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-6">
        {activeTab === 'PARTS' && (
          <PartsHeatmap data={partsData} loading={loadingParts} />
        )}
        {activeTab === 'MERGES' && (
          <MergesMonitor merges={merges} database={selectedDb} table={selectedTable} />
        )}
        {activeTab === 'PIPELINE' && (
          <PipelineGraph onProfileQuery={handleProfileQuery} />
        )}
        {activeTab === 'LOGS' && (
          <QueryLogViewer onSelectQueryForProfile={(_sql) => setActiveTab('PIPELINE')} />
        )}
        {activeTab === 'CODECS' && (
          <CodecOptimizer database={selectedDb} table={selectedTable} />
        )}
        {activeTab === 'LOCAL' && <LocalWorkbench />}
        {activeTab === 'CONSOLE' && (
          <QueryConsole onExecuteQuery={handleExecuteQuery} />
        )}
      </main>

      {/* Command Palette Modal */}
      <CommandPalette
        isOpen={isCommandPaletteOpen}
        onClose={() => setIsCommandPaletteOpen(false)}
        onNavigateTab={handleNavigateFromCommand}
        onOpenReport={() => {
          // Open report via triggering click on header button or state
          const reportBtn = document.querySelector('button[title*="diagnostic audit"]') as HTMLButtonElement;
          if (reportBtn) reportBtn.click();
        }}
        onToggleDemo={() => handleConnect('demo')}
        onSelectTable={(db, tbl) => {
          setSelectedDb(db);
          setSelectedTable(tbl);
        }}
      />

      {/* Subtle Footer */}
      <footer className="border-t border-slate-900 bg-[#070b13] px-6 py-4 text-center text-xs text-slate-500 font-mono">
        ClickLens &bull; Local ClickHouse Observability Studio &bull; Go 1.23 + React 19
      </footer>
    </div>
  );
}

export default App;
