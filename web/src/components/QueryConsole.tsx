import React, { useState, useEffect } from 'react';
import type { ProcessInfo, QueryResult } from '../types';
import { Play, Terminal, Trash2, Clock, RefreshCw } from 'lucide-react';

interface QueryConsoleProps {
  onExecuteQuery: (query: string) => Promise<QueryResult | null>;
}

export const QueryConsole: React.FC<QueryConsoleProps> = ({ onExecuteQuery }) => {
  const [sql, setSql] = useState<string>(
    'SELECT event_type, host, count(), avg(cpu_usage) FROM telemetry.events GROUP BY event_type, host ORDER BY count() DESC LIMIT 50'
  );
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<QueryResult | null>(null);

  const [processes, setProcesses] = useState<ProcessInfo[]>([]);

  const templates = [
    {
      label: 'Sample Query Log',
      sql: 'SELECT event_time, query_duration_ms, memory_usage, query FROM system.query_log ORDER BY event_time DESC LIMIT 20',
    },
    {
      label: 'Parts Summary per Table',
      sql: 'SELECT database, table, count() AS parts_cnt, sum(rows) AS total_rows, formatReadableSize(sum(bytes_on_disk)) AS disk_size FROM system.parts WHERE active = 1 GROUP BY database, table ORDER BY count() DESC LIMIT 15',
    },
    {
      label: 'Column Codec Distribution',
      sql: 'SELECT compression_codec, count() AS columns_count, formatReadableSize(sum(data_compressed_bytes)) AS compressed_size FROM system.columns GROUP BY compression_codec ORDER BY count() DESC',
    },
  ];

  const fetchProcesses = async () => {
    try {
      const res = await fetch('/api/processes');
      if (res.ok) {
        const data: ProcessInfo[] = await res.json();
        setProcesses(data);
      }
    } catch {
      // Ignored
    }
  };

  useEffect(() => {
    fetchProcesses();
    const interval = setInterval(fetchProcesses, 4000);
    return () => clearInterval(interval);
  }, []);

  const handleRun = async () => {
    if (!sql.trim()) return;
    setLoading(true);
    try {
      const res = await onExecuteQuery(sql);
      if (res) {
        setResult(res);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleKill = async (queryId: string) => {
    try {
      const res = await fetch('/api/kill', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query_id: queryId }),
      });
      if (res.ok) {
        fetchProcesses();
      }
    } catch {
      // Ignored
    }
  };

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  return (
    <div className="space-y-6">
      {/* Query Console Editor */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-sky-500/10 text-sky-400 border border-sky-500/20">
              <Terminal className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white">Interactive SQL Console</h3>
              <p className="text-xs text-slate-400">Run queries directly against the ClickHouse cluster</p>
            </div>
          </div>
        </div>

        {/* Templates */}
        <div className="flex items-center gap-2 overflow-x-auto pb-1">
          <span className="text-[11px] text-slate-500 shrink-0">Templates:</span>
          {templates.map((t) => (
            <button
              key={t.label}
              onClick={() => setSql(t.sql)}
              className="text-[11px] px-2.5 py-1 rounded-md bg-slate-950 hover:bg-slate-800 text-slate-400 hover:text-slate-200 border border-slate-800 transition-colors shrink-0 cursor-pointer"
            >
              {t.label}
            </button>
          ))}
        </div>

        <textarea
          value={sql}
          onChange={(e) => setSql(e.target.value)}
          rows={3}
          className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3.5 text-xs text-slate-200 font-mono focus:outline-none focus:border-sky-500 resize-none shadow-inner"
        />

        <div className="flex justify-end">
          <button
            onClick={handleRun}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-sky-400 hover:bg-sky-300 text-slate-950 text-xs font-semibold shadow-lg shadow-sky-500/20 transition-all cursor-pointer disabled:opacity-50"
          >
            {loading ? (
              <div className="w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
            ) : (
              <Play className="w-3.5 h-3.5 fill-current" />
            )}
            <span>Execute Query</span>
          </button>
        </div>

        {/* Query Result Table */}
        {result && (
          <div className="space-y-2 pt-2">
            {result.error ? (
              <div className="p-3 bg-rose-950/40 border border-rose-800/60 rounded-xl text-xs text-rose-300 font-mono">
                {result.error}
              </div>
            ) : (
              <div className="bg-slate-950 rounded-xl border border-slate-800 overflow-hidden shadow-sm">
                <div className="p-2.5 bg-slate-900 border-b border-slate-800 flex items-center justify-between text-xs font-mono text-slate-400">
                  <span>Returned {result.rows_count} rows</span>
                  <span>Execution: {result.execution_time_ms.toFixed(1)} ms</span>
                </div>
                <div className="overflow-x-auto max-h-72">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-slate-950 text-slate-400 border-b border-slate-800 sticky top-0">
                      <tr>
                        {result.columns.map((col) => (
                          <th key={col} className="py-2 px-3 whitespace-nowrap">
                            {col}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {result.rows.map((r, rIdx) => (
                        <tr key={rIdx} className="hover:bg-slate-800/20">
                          {r.map((val, cIdx) => (
                            <td key={cIdx} className="py-2 px-3 whitespace-nowrap text-slate-300">
                              {String(val ?? '')}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Active Running Queries from system.processes */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-emerald-400" />
            <h3 className="text-sm font-semibold text-white">Active Queries (system.processes)</h3>
          </div>
          <button
            onClick={fetchProcesses}
            className="p-1.5 rounded-lg bg-slate-950 hover:bg-slate-800 border border-slate-800 text-slate-400 hover:text-white cursor-pointer"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </button>
        </div>

        {processes.length === 0 ? (
          <div className="p-8 text-center bg-slate-950/40 rounded-xl border border-slate-800/80">
            <p className="text-xs text-slate-500">No active analytical queries currently running.</p>
          </div>
        ) : (
          <div className="divide-y divide-slate-800/80">
            {processes.map((p) => (
              <div key={p.query_id} className="py-3 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
                <div className="space-y-1 max-w-2xl">
                  <div className="flex items-center gap-2 text-xs font-mono">
                    <span className="font-bold text-sky-400">{p.query_id}</span>
                    <span className="text-slate-600">•</span>
                    <span className="text-slate-400">{p.user}@{p.address}</span>
                    <span className="text-slate-600">•</span>
                    <span className="text-amber-400 font-semibold">{p.elapsed_seconds.toFixed(2)}s elapsed</span>
                    <span className="text-slate-600">•</span>
                    <span className="text-slate-300">RAM: {formatBytes(p.memory_usage)}</span>
                  </div>
                  <div className="text-xs font-mono text-slate-300 truncate" title={p.query}>
                    {p.query}
                  </div>
                </div>

                <button
                  onClick={() => handleKill(p.query_id)}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-rose-950/40 hover:bg-rose-900/60 border border-rose-800/60 text-xs font-semibold text-rose-300 transition-colors shrink-0 cursor-pointer"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>Kill Query</span>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
