import React, { useState } from 'react';
import type { DatabaseInfo, ServerStatus } from '../types';
import { Activity, Database, HardDrive, Layers, RefreshCw, Zap, Server, ChevronDown } from 'lucide-react';

interface HeaderProps {
  status: ServerStatus | null;
  databases: DatabaseInfo[];
  selectedDb: string;
  selectedTable: string;
  onSelectDb: (db: string) => void;
  onSelectTable: (table: string) => void;
  onConnect: (dsn: string) => void;
  onRefresh: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  status,
  databases,
  selectedDb,
  selectedTable,
  onSelectDb,
  onSelectTable,
  onConnect,
  onRefresh,
}) => {
  const [showConnectModal, setShowConnectModal] = useState(false);
  const [dsnInput, setDsnInput] = useState(status?.current_dsn || '');

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const currentDbObj = databases.find((d) => d.name === selectedDb);
  const availableTables = currentDbObj?.tables || [];

  const handleConnectSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConnect(dsnInput);
    setShowConnectModal(false);
  };

  return (
    <header className="border-b border-ch-border bg-[#0b1120] text-slate-200 px-6 py-3.5 sticky top-0 z-30 shadow-md">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        {/* Brand & Connection State */}
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-amber-500 to-sky-500 flex items-center justify-center shadow-lg shadow-sky-500/20 ring-1 ring-white/20">
            <span className="font-black text-xl text-slate-950 tracking-tighter">CL</span>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="font-bold text-lg text-white tracking-tight">ClickLens</h1>
              <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-slate-800 text-slate-400 border border-slate-700">
                v1.0
              </span>
            </div>
            <p className="text-xs text-slate-400 font-mono">
              ClickHouse Observability Studio
            </p>
          </div>

          <div className="h-6 w-px bg-slate-800 mx-2 hidden sm:block" />

          {/* Connection Status Pill */}
          <button
            onClick={() => setShowConnectModal(true)}
            className={`flex items-center gap-2 text-xs px-3 py-1.5 rounded-lg border transition-all cursor-pointer ${
              status?.is_demo
                ? 'bg-amber-950/40 text-amber-300 border-amber-800/60 hover:bg-amber-900/50'
                : status?.connected
                ? 'bg-emerald-950/40 text-emerald-300 border-emerald-800/60 hover:bg-emerald-900/50'
                : 'bg-rose-950/40 text-rose-300 border-rose-800/60 hover:bg-rose-900/50'
            }`}
          >
            <div
              className={`w-2 h-2 rounded-full animate-pulse ${
                status?.is_demo
                  ? 'bg-amber-400'
                  : status?.connected
                  ? 'bg-emerald-400'
                  : 'bg-rose-400'
              }`}
            />
            <span className="font-semibold">
              {status?.is_demo ? 'Simulation Mode' : status?.connected ? 'Live ClickHouse' : 'Disconnected'}
            </span>
            <ChevronDown className="w-3.5 h-3.5 opacity-60" />
          </button>
        </div>

        {/* Database & Table Selectors */}
        <div className="flex items-center gap-2.5 w-full md:w-auto">
          <div className="flex items-center gap-2 bg-slate-900/80 px-3 py-1.5 rounded-lg border border-slate-800">
            <Database className="w-3.5 h-3.5 text-sky-400" />
            <select
              value={selectedDb}
              onChange={(e) => onSelectDb(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium"
            >
              {databases.map((db) => (
                <option key={db.name} value={db.name} className="bg-slate-900 text-slate-200">
                  {db.name} ({db.table_count})
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-2 bg-slate-900/80 px-3 py-1.5 rounded-lg border border-slate-800">
            <Layers className="w-3.5 h-3.5 text-amber-400" />
            <select
              value={selectedTable}
              onChange={(e) => onSelectTable(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium"
            >
              {availableTables.map((tbl) => (
                <option key={tbl} value={tbl} className="bg-slate-900 text-slate-200">
                  {tbl}
                </option>
              ))}
            </select>
          </div>

          <button
            onClick={onRefresh}
            title="Refresh Data"
            className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-400 hover:text-white transition-colors cursor-pointer"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>

        {/* Quick Node Telemetry Stats */}
        <div className="hidden lg:flex items-center gap-4 text-xs font-mono">
          <div className="flex items-center gap-1.5 text-slate-300">
            <Zap className="w-3.5 h-3.5 text-amber-400" />
            <span className="text-slate-500">Merges:</span>
            <span className="font-semibold text-white">{status?.active_merges ?? 0}</span>
          </div>
          <div className="flex items-center gap-1.5 text-slate-300">
            <Activity className="w-3.5 h-3.5 text-emerald-400" />
            <span className="text-slate-500">Queries:</span>
            <span className="font-semibold text-white">{status?.active_queries ?? 0}</span>
          </div>
          <div className="flex items-center gap-1.5 text-slate-300">
            <HardDrive className="w-3.5 h-3.5 text-sky-400" />
            <span className="text-slate-500">Storage:</span>
            <span className="font-semibold text-white">{formatBytes(status?.total_bytes_disk ?? 0)}</span>
          </div>
        </div>
      </div>

      {/* Connection Modal */}
      {showConnectModal && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-2.5 rounded-xl bg-sky-500/10 text-sky-400 border border-sky-500/20">
                <Server className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-white">ClickHouse Connection</h3>
                <p className="text-xs text-slate-400">Configure connection DSN or toggle simulation mode</p>
              </div>
            </div>

            <form onSubmit={handleConnectSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Connection DSN (Native TCP or HTTP)
                </label>
                <input
                  type="text"
                  value={dsnInput}
                  onChange={(e) => setDsnInput(e.target.value)}
                  placeholder="clickhouse://default:password@localhost:9000/default"
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3.5 py-2.5 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-sky-500 font-mono"
                />
                <p className="text-[11px] text-slate-500 mt-1.5">
                  Enter <code className="text-amber-400">demo</code> to activate the realistic in-memory telemetry simulation.
                </p>
              </div>

              <div className="p-3 bg-slate-950/50 rounded-lg border border-slate-800/80 text-xs space-y-1">
                <div className="flex justify-between text-slate-400">
                  <span>Server Version:</span>
                  <span className="font-mono text-slate-200">{status?.version || 'N/A'}</span>
                </div>
                <div className="flex justify-between text-slate-400">
                  <span>Protocol:</span>
                  <span className="font-mono text-slate-200">{status?.protocol || 'Native TCP'}</span>
                </div>
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowConnectModal(false)}
                  className="px-4 py-2 text-xs text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setDsnInput('demo');
                    onConnect('demo');
                    setShowConnectModal(false);
                  }}
                  className="px-4 py-2 text-xs font-medium text-amber-400 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/30 rounded-lg transition-colors cursor-pointer"
                >
                  Use Demo Mode
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-xs font-semibold text-slate-950 bg-sky-400 hover:bg-sky-300 rounded-lg shadow-lg shadow-sky-500/20 transition-colors cursor-pointer"
                >
                  Connect
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </header>
  );
};
