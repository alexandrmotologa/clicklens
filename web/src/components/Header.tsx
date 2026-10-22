import React, { useState } from 'react';
import type { DatabaseInfo, ServerStatus, HealthAuditReport } from '../types';
import { 
  Activity, 
  Database, 
  HardDrive, 
  Layers, 
  RefreshCw, 
  Zap, 
  Server, 
  ChevronDown, 
  Command, 
  Download, 
  Copy, 
  Check, 
  X,
  ShieldCheck
} from 'lucide-react';

interface HeaderProps {
  status: ServerStatus | null;
  databases: DatabaseInfo[];
  selectedDb: string;
  selectedTable: string;
  onSelectDb: (db: string) => void;
  onSelectTable: (table: string) => void;
  onConnect: (dsn: string) => void;
  onRefresh: () => void;
  onOpenCommandPalette: () => void;
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
  onOpenCommandPalette,
}) => {
  const [showConnectModal, setShowConnectModal] = useState(false);
  const [dsnInput, setDsnInput] = useState(status?.current_dsn || '');

  // Diagnostic Report Modal state
  const [showReportModal, setShowReportModal] = useState(false);
  const [reportLoading, setReportLoading] = useState(false);
  const [reportData, setReportData] = useState<HealthAuditReport | null>(null);
  const [copiedReport, setCopiedReport] = useState(false);

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

  const handleOpenReport = async () => {
    setShowReportModal(true);
    setReportLoading(true);
    try {
      const res = await fetch(`/api/report?database=${selectedDb}&table=${selectedTable}`);
      if (res.ok) {
        const data: HealthAuditReport = await res.json();
        setReportData(data);
      }
    } catch (e) {
      console.error('Failed to generate diagnostic report', e);
    } finally {
      setReportLoading(false);
    }
  };

  const downloadFile = (content: string, filename: string, type: string) => {
    const blob = new Blob([content], { type });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleCopyMarkdown = (md: string) => {
    navigator.clipboard.writeText(md);
    setCopiedReport(true);
    setTimeout(() => setCopiedReport(false), 2000);
  };

  return (
    <header className="border-b border-[#1f2430] bg-[#0b0e14] text-slate-200 px-6 py-3.5 sticky top-0 z-30 shadow-md">
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
                v1.1
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

        {/* Database & Table Selectors + Actions */}
        <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
          <div className="flex items-center gap-2 bg-slate-900/80 px-2.5 py-1.5 rounded-lg border border-slate-800 flex-1 sm:flex-none">
            <Database className="w-3.5 h-3.5 text-sky-400 shrink-0" />
            <select
              value={selectedDb}
              onChange={(e) => onSelectDb(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium w-full sm:w-auto"
            >
              {databases.map((db) => (
                <option key={db.name} value={db.name} className="bg-slate-900 text-slate-200">
                  {db.name} ({db.table_count})
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-2 bg-slate-900/80 px-2.5 py-1.5 rounded-lg border border-slate-800 flex-1 sm:flex-none">
            <Layers className="w-3.5 h-3.5 text-amber-400 shrink-0" />
            <select
              value={selectedTable}
              onChange={(e) => onSelectTable(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium w-full sm:w-auto"
            >
              {availableTables.map((tbl) => (
                <option key={tbl} value={tbl} className="bg-slate-900 text-slate-200">
                  {tbl}
                </option>
              ))}
            </select>
          </div>

          {/* Diagnostic Report Export Button */}
          <button
            onClick={handleOpenReport}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/30 text-xs font-medium transition-colors cursor-pointer shrink-0"
            title="Generate and export diagnostic audit report"
          >
            <ShieldCheck className="w-3.5 h-3.5 text-rose-400" />
            <span className="inline">Report</span>
          </button>

          {/* Command Palette Trigger */}
          <button
            onClick={onOpenCommandPalette}
            className="hidden sm:flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-400 hover:text-slate-200 text-xs font-mono transition-colors cursor-pointer shrink-0"
            title="Open Command Palette (Ctrl+K)"
          >
            <Command className="w-3.5 h-3.5" />
            <span className="text-[11px]">Ctrl+K</span>
          </button>

          <button
            onClick={onRefresh}
            title="Refresh Data"
            className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-400 hover:text-white transition-colors cursor-pointer shrink-0"
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

      {/* Health Audit Diagnostic Report Modal */}
      {showReportModal && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#12151e] border border-[#232938] rounded-2xl max-w-3xl w-full p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#1f2533] pb-3">
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-rose-500/10 text-rose-400 border border-rose-500/20">
                  <ShieldCheck className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white">Cluster Diagnostic & Health Audit Report</h3>
                  <p className="text-xs text-slate-400">Target: {selectedDb}.{selectedTable}</p>
                </div>
              </div>
              <button
                onClick={() => setShowReportModal(false)}
                className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {reportLoading ? (
              <div className="p-12 text-center text-slate-400">
                <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-rose-400" />
                <p className="text-xs font-mono">Compiling ClickHouse diagnostic audit report...</p>
              </div>
            ) : reportData ? (
              <div className="space-y-4">
                {/* Metric Summary Badges */}
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                  <div className="p-3 bg-[#161a25] rounded-xl border border-[#232938]">
                    <div className="text-slate-400">Overall Rating</div>
                    <div className="text-base font-bold text-emerald-400 mt-0.5">{reportData.overall_health}</div>
                  </div>
                  <div className="p-3 bg-[#161a25] rounded-xl border border-[#232938]">
                    <div className="text-slate-400">Active Parts</div>
                    <div className="text-base font-bold text-white font-mono mt-0.5">{reportData.active_parts}</div>
                  </div>
                  <div className="p-3 bg-[#161a25] rounded-xl border border-[#232938]">
                    <div className="text-slate-400">Compression</div>
                    <div className="text-base font-bold text-sky-400 font-mono mt-0.5">{reportData.overall_ratio.toFixed(2)}x</div>
                  </div>
                  <div className="p-3 bg-[#161a25] rounded-xl border border-[#232938]">
                    <div className="text-slate-400">Potential Savings</div>
                    <div className="text-base font-bold text-amber-400 font-mono mt-0.5">{formatBytes(reportData.potential_savings_bytes)}</div>
                  </div>
                </div>

                {/* Markdown Preview Area */}
                <pre className="p-4 bg-[#0a0c10] border border-[#1e2330] rounded-xl text-xs font-mono text-slate-300 max-h-72 overflow-y-auto whitespace-pre-wrap leading-relaxed">
                  {reportData.markdown_content}
                </pre>

                {/* Action Buttons */}
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pt-2">
                  <span className="text-[11px] text-slate-500 font-mono">Ready for export to SRE / DBA documentation</span>
                  <div className="flex items-center gap-2 w-full sm:w-auto">
                    <button
                      onClick={() => handleCopyMarkdown(reportData.markdown_content)}
                      className="flex-1 sm:flex-none px-3.5 py-2 bg-[#1a1f2c] hover:bg-[#23293b] border border-[#2e374d] text-slate-300 rounded-lg text-xs font-medium flex items-center justify-center gap-1.5 transition-colors cursor-pointer"
                    >
                      {copiedReport ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                      <span>{copiedReport ? 'Copied' : 'Copy MD'}</span>
                    </button>
                    <button
                      onClick={() => downloadFile(reportData.markdown_content, `clicklens_report_${selectedDb}_${selectedTable}.md`, 'text/markdown')}
                      className="flex-1 sm:flex-none px-3.5 py-2 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-xs font-semibold flex items-center justify-center gap-1.5 transition-colors shadow-sm cursor-pointer"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>Download .MD</span>
                    </button>
                    <button
                      onClick={() => downloadFile(JSON.stringify(reportData, null, 2), `clicklens_report_${selectedDb}_${selectedTable}.json`, 'application/json')}
                      className="flex-1 sm:flex-none px-3.5 py-2 bg-[#1a1f2c] hover:bg-[#23293b] border border-[#2e374d] text-slate-300 rounded-lg text-xs font-medium flex items-center justify-center gap-1.5 transition-colors cursor-pointer"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>Download .JSON</span>
                    </button>
                  </div>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      )}

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
