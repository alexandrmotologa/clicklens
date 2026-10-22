import React, { useState, useEffect } from 'react';
import type { QueryLogEntry, QueryLogResponse } from '../types';
import { 
  Clock, 
  Cpu, 
  HardDrive, 
  AlertTriangle, 
  Search, 
  RefreshCw, 
  Copy, 
  Check, 
  Play, 
  Activity,
  ArrowUpDown
} from 'lucide-react';

interface QueryLogViewerProps {
  onSelectQueryForProfile?: (sql: string) => void;
}

export const QueryLogViewer: React.FC<QueryLogViewerProps> = ({ onSelectQueryForProfile }) => {
  const [filter, setFilter] = useState<'slow' | 'memory' | 'spill' | 'errors'>('slow');
  const [entries, setEntries] = useState<QueryLogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const fetchLogs = async (currentFilter: string) => {
    setLoading(true);
    try {
      const res = await fetch(`/api/query-log?filter=${currentFilter}&limit=40`);
      if (res.ok) {
        const data: QueryLogResponse = await res.json();
        setEntries(data.entries || []);
      }
    } catch (e) {
      console.error('Failed to load query log', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs(filter);
  }, [filter]);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const filteredEntries = entries.filter((e) => {
    if (!searchQuery) return true;
    const lower = searchQuery.toLowerCase();
    return (
      e.query_id.toLowerCase().includes(lower) ||
      e.query_text.toLowerCase().includes(lower) ||
      (e.exception && e.exception.toLowerCase().includes(lower))
    );
  });

  const formatBytes = (bytes: number): string => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms.toFixed(1)} ms`;
    return `${(ms / 1000).toFixed(2)} s`;
  };

  return (
    <div className="space-y-6">
      {/* Top Controls & Category Filters */}
      <div className="bg-[#12141a] border border-[#212631] rounded-xl p-5 shadow-sm">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-white flex items-center gap-2">
              <Activity className="w-5 h-5 text-indigo-400" />
              Historical Query Telemetry
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              Live introspection into system.query_log: analyze slow queries, peak RAM spikes, disk spills, and exceptions.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => fetchLogs(filter)}
              disabled={loading}
              className="px-3 py-1.5 bg-[#1a1f2c] hover:bg-[#23293b] border border-[#2e374d] text-slate-300 text-xs rounded-lg transition-colors flex items-center gap-1.5"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-indigo-400' : ''}`} />
              Refresh
            </button>
          </div>
        </div>

        {/* Filter Pills */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-5">
          <button
            onClick={() => setFilter('slow')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg border text-sm font-medium transition-all ${
              filter === 'slow'
                ? 'bg-amber-500/10 border-amber-500/30 text-amber-300 shadow-sm'
                : 'bg-[#181c26] border-[#252c3d] text-slate-400 hover:text-slate-200'
            }`}
          >
            <Clock className="w-4 h-4 text-amber-400" />
            <span>Slow Queries</span>
          </button>

          <button
            onClick={() => setFilter('memory')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg border text-sm font-medium transition-all ${
              filter === 'memory'
                ? 'bg-purple-500/10 border-purple-500/30 text-purple-300 shadow-sm'
                : 'bg-[#181c26] border-[#252c3d] text-slate-400 hover:text-slate-200'
            }`}
          >
            <Cpu className="w-4 h-4 text-purple-400" />
            <span>Memory Spenders</span>
          </button>

          <button
            onClick={() => setFilter('spill')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg border text-sm font-medium transition-all ${
              filter === 'spill'
                ? 'bg-blue-500/10 border-blue-500/30 text-blue-300 shadow-sm'
                : 'bg-[#181c26] border-[#252c3d] text-slate-400 hover:text-slate-200'
            }`}
          >
            <HardDrive className="w-4 h-4 text-blue-400" />
            <span>Disk Spills</span>
          </button>

          <button
            onClick={() => setFilter('errors')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg border text-sm font-medium transition-all ${
              filter === 'errors'
                ? 'bg-rose-500/10 border-rose-500/30 text-rose-300 shadow-sm'
                : 'bg-[#181c26] border-[#252c3d] text-slate-400 hover:text-slate-200'
            }`}
          >
            <AlertTriangle className="w-4 h-4 text-rose-400" />
            <span>Errors & OOMs</span>
          </button>
        </div>

        {/* Search input */}
        <div className="mt-4 relative">
          <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search query text, query ID, or error message..."
            className="w-full bg-[#181c26] border border-[#252c3d] rounded-lg pl-9 pr-4 py-2 text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
          />
        </div>
      </div>

      {/* Query Entries List */}
      <div className="space-y-3">
        {loading && entries.length === 0 ? (
          <div className="p-12 text-center text-slate-500 bg-[#12141a] border border-[#212631] rounded-xl">
            <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-indigo-400" />
            Loading query log entries...
          </div>
        ) : filteredEntries.length === 0 ? (
          <div className="p-12 text-center text-slate-500 bg-[#12141a] border border-[#212631] rounded-xl">
            No queries matched the current filter or search criteria.
          </div>
        ) : (
          filteredEntries.map((entry) => {
            const isExpanded = expandedId === entry.query_id;
            const isError = entry.exception_code !== 0 || entry.type === 'ExceptionWhileProcessing';

            return (
              <div
                key={entry.query_id}
                className={`bg-[#12141a] border rounded-xl overflow-hidden transition-all ${
                  isError ? 'border-rose-900/40 hover:border-rose-700/60' : 'border-[#212631] hover:border-[#2f374a]'
                }`}
              >
                {/* Entry Header */}
                <div
                  onClick={() => setExpandedId(isExpanded ? null : entry.query_id)}
                  className="p-4 cursor-pointer flex flex-col md:flex-row md:items-center justify-between gap-3 select-none"
                >
                  <div className="flex items-start gap-3">
                    <div
                      className={`p-2 rounded-lg mt-0.5 ${
                        isError
                          ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                          : entry.has_spill
                          ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                          : 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
                      }`}
                    >
                      {isError ? (
                        <AlertTriangle className="w-4 h-4" />
                      ) : entry.has_spill ? (
                        <HardDrive className="w-4 h-4" />
                      ) : (
                        <Clock className="w-4 h-4" />
                      )}
                    </div>

                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs font-semibold text-slate-300">
                          {entry.query_id}
                        </span>
                        {entry.has_spill && (
                          <span className="px-1.5 py-0.5 bg-amber-500/20 border border-amber-500/40 text-amber-300 text-[10px] font-medium rounded">
                            DISK SPILL
                          </span>
                        )}
                        {isError && (
                          <span className="px-1.5 py-0.5 bg-rose-500/20 border border-rose-500/40 text-rose-300 text-[10px] font-medium rounded">
                            CODE {entry.exception_code}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-400 font-mono mt-1 line-clamp-1 max-w-2xl">
                        {entry.query_text}
                      </p>
                    </div>
                  </div>

                  {/* Metrics Badges */}
                  <div className="flex items-center gap-4 text-xs font-mono text-slate-400">
                    <div className="flex items-center gap-1.5 bg-[#181c26] px-2.5 py-1.5 rounded-lg border border-[#252c3d]">
                      <Clock className="w-3.5 h-3.5 text-amber-400" />
                      <span className="text-slate-200">{formatDuration(entry.duration_ms)}</span>
                    </div>

                    <div className="flex items-center gap-1.5 bg-[#181c26] px-2.5 py-1.5 rounded-lg border border-[#252c3d]">
                      <Cpu className="w-3.5 h-3.5 text-purple-400" />
                      <span className="text-slate-200">{formatBytes(entry.memory_usage)}</span>
                    </div>

                    <div className="hidden lg:flex items-center gap-1.5 bg-[#181c26] px-2.5 py-1.5 rounded-lg border border-[#252c3d]">
                      <ArrowUpDown className="w-3.5 h-3.5 text-blue-400" />
                      <span className="text-slate-200">{entry.read_rows.toLocaleString()} rows</span>
                    </div>
                  </div>
                </div>

                {/* Expanded Details */}
                {isExpanded && (
                  <div className="border-t border-[#1e232f] bg-[#0d0e12] p-4 space-y-4">
                    {/* Error Box if any */}
                    {entry.exception && (
                      <div className="p-3 bg-rose-950/30 border border-rose-800/40 rounded-lg text-rose-300 text-xs font-mono">
                        <div className="font-semibold mb-1 flex items-center gap-1.5">
                          <AlertTriangle className="w-3.5 h-3.5 text-rose-400" />
                          ClickHouse Exception:
                        </div>
                        {entry.exception}
                      </div>
                    )}

                    {/* Full SQL Block */}
                    <div>
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                          Full Query SQL
                        </span>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => copyToClipboard(entry.query_text, entry.query_id)}
                            className="text-xs px-2.5 py-1 bg-[#1a1f2c] hover:bg-[#23293b] border border-[#2e374d] text-slate-300 rounded flex items-center gap-1 transition-colors"
                          >
                            {copiedId === entry.query_id ? (
                              <>
                                <Check className="w-3 h-3 text-emerald-400" />
                                <span className="text-emerald-400">Copied</span>
                              </>
                            ) : (
                              <>
                                <Copy className="w-3 h-3 text-slate-400" />
                                <span>Copy SQL</span>
                              </>
                            )}
                          </button>

                          {onSelectQueryForProfile && (
                            <button
                              onClick={() => onSelectQueryForProfile(entry.query_text)}
                              className="text-xs px-2.5 py-1 bg-indigo-600/20 hover:bg-indigo-600/30 border border-indigo-500/40 text-indigo-300 rounded flex items-center gap-1 transition-colors"
                            >
                              <Play className="w-3 h-3 text-indigo-400" />
                              <span>Profile in Pipeline</span>
                            </button>
                          )}
                        </div>
                      </div>

                      <pre className="p-3 bg-[#0a0a0d] border border-[#1e232f] rounded-lg text-xs font-mono text-slate-300 overflow-x-auto whitespace-pre-wrap">
                        {entry.query_text}
                      </pre>
                    </div>

                    {/* Metrics Breakdown Grid */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                      <div className="p-2.5 bg-[#141720] border border-[#1f2533] rounded-lg">
                        <div className="text-slate-400">Event Time</div>
                        <div className="text-white font-mono mt-0.5">
                          {new Date(entry.event_time).toLocaleTimeString()}
                        </div>
                      </div>

                      <div className="p-2.5 bg-[#141720] border border-[#1f2533] rounded-lg">
                        <div className="text-slate-400">Read Bytes</div>
                        <div className="text-white font-mono mt-0.5">
                          {formatBytes(entry.read_bytes)}
                        </div>
                      </div>

                      <div className="p-2.5 bg-[#141720] border border-[#1f2533] rounded-lg">
                        <div className="text-slate-400">Result Rows</div>
                        <div className="text-white font-mono mt-0.5">
                          {entry.result_rows.toLocaleString()}
                        </div>
                      </div>

                      <div className="p-2.5 bg-[#141720] border border-[#1f2533] rounded-lg">
                        <div className="text-slate-400">Result Size</div>
                        <div className="text-white font-mono mt-0.5">
                          {formatBytes(entry.result_bytes)}
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
