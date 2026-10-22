import React, { useState, useEffect } from 'react';
import type { LocalFileProfile, QueryResult } from '../types';
import { Database, FileCode, Play, Search, Terminal } from 'lucide-react';

export const LocalWorkbench: React.FC<LocalWorkbenchProps> = () => {
  const [filePath, setFilePath] = useState<string>('./testdata/sample_telemetry.csv');
  const [loading, setLoading] = useState<boolean>(false);
  const [fileProfile, setFileProfile] = useState<LocalFileProfile | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [localQuery, setLocalQuery] = useState<string>(
    "SELECT service_name, count(), avg(cpu_percent) FROM file('./testdata/sample_telemetry.csv', CSVWithNames) GROUP BY service_name"
  );
  const [queryLoading, setQueryLoading] = useState<boolean>(false);
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null);
  const [runnerStatus, setRunnerStatus] = useState<{ available: boolean; binary_path: string } | null>(null);

  useEffect(() => {
    async function checkRunner() {
      try {
        const res = await fetch('/api/local/status');
        if (res.ok) {
          const data = await res.json();
          setRunnerStatus(data);
        }
      } catch {
        // Ignored
      }
    }
    checkRunner();
  }, []);

  const handleInspect = async (pathOverride?: string) => {
    const targetPath = pathOverride || filePath;
    if (!targetPath.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/local/inspect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ file_path: targetPath }),
      });
      if (res.ok) {
        const data: LocalFileProfile = await res.json();
        setFileProfile(data);
      } else {
        const errData = await res.json();
        setError(errData.error || 'Failed to inspect file');
      }
    } catch (e: any) {
      setError(e.message || 'Inspection failed');
    } finally {
      setLoading(false);
    }
  };

  const handleRunLocalQuery = async () => {
    if (!localQuery.trim()) return;
    setQueryLoading(true);
    try {
      const res = await fetch('/api/local/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: localQuery }),
      });
      const data: QueryResult = await res.json();
      setQueryResult(data);
    } catch (e: any) {
      setQueryResult({
        query: localQuery,
        columns: [],
        rows: [],
        rows_count: 0,
        execution_time_ms: 0,
        error: e.message || 'Execution error',
      });
    } finally {
      setQueryLoading(false);
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
      {/* Engine Status Banner */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <Terminal className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white">Serverless Local File Workbench</h3>
              <p className="text-xs text-slate-400">
                Inspect and query Parquet, CSV, and TSV files locally without running a ClickHouse server daemon
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 text-xs font-mono">
            <span className="text-slate-500">Runner:</span>
            {runnerStatus?.available ? (
              <span className="px-2.5 py-0.5 rounded-full bg-emerald-950 text-emerald-300 border border-emerald-800 font-semibold">
                clickhouse-local active
              </span>
            ) : (
              <span className="px-2.5 py-0.5 rounded-full bg-sky-950 text-sky-300 border border-sky-800 font-semibold">
                pure Go engine fallback
              </span>
            )}
          </div>
        </div>

        {/* File Path Bar */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
          <div className="relative flex-1">
            <input
              type="text"
              value={filePath}
              onChange={(e) => setFilePath(e.target.value)}
              placeholder="e.g. ./testdata/sample_telemetry.csv or C:/datasets/analytics.parquet"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-xs text-slate-200 font-mono focus:outline-none focus:border-sky-500 placeholder-slate-600 shadow-inner"
            />
          </div>
          <button
            onClick={() => handleInspect()}
            disabled={loading}
            className="flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl bg-sky-400 hover:bg-sky-300 text-slate-950 text-xs font-semibold shadow-lg shadow-sky-500/20 transition-all cursor-pointer disabled:opacity-50"
          >
            {loading ? (
              <div className="w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
            ) : (
              <Search className="w-3.5 h-3.5" />
            )}
            <span>Inspect File</span>
          </button>
        </div>

        {error && (
          <div className="p-3 bg-rose-950/40 border border-rose-800/60 rounded-xl text-xs text-rose-300">
            {error}
          </div>
        )}
      </div>

      {/* File Inspection Result View */}
      {fileProfile && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Detected Format</span>
              <div className="text-lg font-bold text-sky-400 font-mono mt-0.5">{fileProfile.format}</div>
            </div>
            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">File Size</span>
              <div className="text-lg font-bold text-white font-mono mt-0.5">
                {formatBytes(fileProfile.file_size_bytes)}
              </div>
            </div>
            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Row Count</span>
              <div className="text-lg font-bold text-white font-mono mt-0.5">
                {fileProfile.row_count.toLocaleString()}
              </div>
            </div>
            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Inspector Engine</span>
              <div className="text-xs font-bold text-emerald-400 font-mono mt-1 truncate">
                {fileProfile.runner_used}
              </div>
            </div>
          </div>

          {/* Schema Columns Table */}
          <div className="bg-slate-900/80 rounded-2xl border border-slate-800 overflow-hidden shadow-sm">
            <div className="p-3.5 border-b border-slate-800 flex items-center gap-2">
              <FileCode className="w-4 h-4 text-sky-400" />
              <h4 className="text-xs font-bold text-white font-mono">Inferred Column Schema ({fileProfile.columns.length})</h4>
            </div>
            <div className="overflow-x-auto max-h-64">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-950 text-slate-400 border-b border-slate-800 font-mono sticky top-0">
                  <tr>
                    <th className="py-2 px-3">Column Name</th>
                    <th className="py-2 px-3">Inferred Type</th>
                    <th className="py-2 px-3">Nullable</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 font-mono">
                  {fileProfile.columns.map((c) => (
                    <tr key={c.name} className="hover:bg-slate-800/20">
                      <td className="py-2 px-3 font-semibold text-white">{c.name}</td>
                      <td className="py-2 px-3 text-sky-300">{c.type}</td>
                      <td className="py-2 px-3 text-slate-400">{c.nullable ? 'Yes' : 'No'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Sample Rows Table */}
          {fileProfile.sample_rows && fileProfile.sample_rows.length > 0 && (
            <div className="bg-slate-900/80 rounded-2xl border border-slate-800 overflow-hidden shadow-sm">
              <div className="p-3.5 border-b border-slate-800 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Database className="w-4 h-4 text-amber-400" />
                  <h4 className="text-xs font-bold text-white font-mono">
                    Sample Data Preview (First {fileProfile.sample_rows.length} rows)
                  </h4>
                </div>
              </div>
              <div className="overflow-x-auto max-h-64">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-950 text-slate-400 border-b border-slate-800 font-mono sticky top-0">
                    <tr>
                      {fileProfile.columns.map((c) => (
                        <th key={c.name} className="py-2 px-3 whitespace-nowrap">
                          {c.name}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60 font-mono">
                    {fileProfile.sample_rows.map((row, idx) => (
                      <tr key={idx} className="hover:bg-slate-800/20">
                        {fileProfile.columns.map((c) => (
                          <td key={c.name} className="py-2 px-3 whitespace-nowrap text-slate-300">
                            {String(row[c.name] ?? '')}
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

      {/* Ad-Hoc SQL Runner for clickhouse-local */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Play className="w-4 h-4 text-sky-400" />
            <h4 className="text-xs font-bold text-white font-mono">Ad-Hoc Query on Local File</h4>
          </div>
        </div>

        <textarea
          value={localQuery}
          onChange={(e) => setLocalQuery(e.target.value)}
          rows={3}
          className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3.5 text-xs text-slate-200 font-mono focus:outline-none focus:border-sky-500 resize-none shadow-inner"
        />

        <div className="flex justify-end">
          <button
            onClick={handleRunLocalQuery}
            disabled={queryLoading}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-sky-400 hover:bg-sky-300 text-slate-950 text-xs font-semibold shadow-lg shadow-sky-500/20 transition-all cursor-pointer disabled:opacity-50"
          >
            {queryLoading ? (
              <div className="w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
            ) : (
              <Play className="w-3.5 h-3.5 fill-current" />
            )}
            <span>Execute via clickhouse-local</span>
          </button>
        </div>

        {queryResult && (
          <div className="space-y-2 pt-2">
            {queryResult.error ? (
              <div className="p-3 bg-rose-950/40 border border-rose-800/60 rounded-xl text-xs text-rose-300 font-mono">
                {queryResult.error}
              </div>
            ) : (
              <div className="bg-slate-950 rounded-xl border border-slate-800 overflow-hidden shadow-sm">
                <div className="p-2.5 bg-slate-900 border-b border-slate-800 flex items-center justify-between text-xs font-mono text-slate-400">
                  <span>Returned {queryResult.rows_count} rows</span>
                  <span>Execution: {queryResult.execution_time_ms.toFixed(1)} ms</span>
                </div>
                <div className="overflow-x-auto max-h-64">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-slate-950 text-slate-400 border-b border-slate-800 sticky top-0">
                      <tr>
                        {queryResult.columns.map((col) => (
                          <th key={col} className="py-2 px-3 whitespace-nowrap">
                            {col}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {queryResult.rows.map((r, rIdx) => (
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
    </div>
  );
};

export interface LocalWorkbenchProps {}
