import React, { useState, useEffect } from 'react';
import type { MergeInfo, MutationInfo } from '../types';
import { Activity, Clock, Flame, GitMerge } from 'lucide-react';

interface MergesMonitorProps {
  merges: MergeInfo[];
  database: string;
  table: string;
}

export const MergesMonitor: React.FC<MergesMonitorProps> = ({ merges, database, table }) => {
  const [mutations, setMutations] = useState<MutationInfo[]>([]);

  useEffect(() => {
    async function loadMutations() {
      try {
        const res = await fetch(`/api/mutations?database=${database}&table=${table}`);
        if (res.ok) {
          const data: MutationInfo[] = await res.json();
          setMutations(data);
        }
      } catch {
        // Ignored
      }
    }
    loadMutations();
    const interval = setInterval(loadMutations, 5000);
    return () => clearInterval(interval);
  }, [database, table]);

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  return (
    <div className="space-y-6">
      {/* Active Background Merges Cockpit */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <Flame className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white">Active Background Merges</h3>
              <p className="text-xs text-slate-400">
                Live stream from <code className="text-sky-400">system.merges</code>
              </p>
            </div>
          </div>
          <span className="flex items-center gap-2 text-xs font-mono px-3 py-1 rounded-full bg-slate-950 border border-slate-800 text-slate-400">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
            <span>{merges.length} Active</span>
          </span>
        </div>

        {merges.length === 0 ? (
          <div className="p-10 text-center bg-slate-950/40 rounded-xl border border-slate-800/80">
            <GitMerge className="w-8 h-8 text-slate-600 mx-auto mb-2" />
            <p className="text-xs text-slate-400">No background merges currently running.</p>
            <p className="text-[11px] text-slate-500 mt-0.5">
              ClickHouse triggers merges asynchronously when small parts accumulate.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3.5">
            {merges.map((m, idx) => {
              const readMB = (m.read_throughput_bytes_sec / (1024 * 1024)).toFixed(1);
              const writeMB = (m.write_throughput_bytes_sec / (1024 * 1024)).toFixed(1);

              return (
                <div
                  key={`${m.database}-${m.table}-${m.result_part_name}-${idx}`}
                  className="p-4 bg-slate-950/80 rounded-xl border border-slate-800/90 space-y-3 shadow-inner"
                >
                  <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-mono font-bold text-sky-400">
                        {m.database}.{m.table}
                      </span>
                      <span className="text-slate-600">•</span>
                      <span className="text-xs text-slate-400">
                        Merging <strong className="text-white">{m.num_parts}</strong> parts into{' '}
                        <code className="text-amber-300 font-mono">{m.result_part_name}</code>
                      </span>
                    </div>

                    <div className="flex items-center gap-3 text-xs font-mono">
                      <span className="text-slate-400 flex items-center gap-1">
                        <Clock className="w-3.5 h-3.5 text-slate-500" />
                        {m.elapsed_seconds.toFixed(1)}s
                      </span>
                      <span className="text-amber-400 font-bold">{m.progress.toFixed(1)}%</span>
                    </div>
                  </div>

                  {/* Progress Bar */}
                  <div className="h-2.5 w-full bg-slate-900 rounded-full overflow-hidden border border-slate-800 relative">
                    <div
                      className="h-full bg-gradient-to-r from-amber-500 to-sky-500 rounded-full transition-all duration-300 relative"
                      style={{ width: `${Math.min(100, Math.max(3, m.progress))}%` }}
                    >
                      <div className="absolute inset-0 bg-white/20 animate-pulse" />
                    </div>
                  </div>

                  {/* Throughput & Memory Stats */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 pt-1 text-xs font-mono text-slate-400">
                    <div>
                      <span className="text-[11px] text-slate-500 block">Read Speed:</span>
                      <span className="text-slate-200">{readMB} MB/s</span>
                    </div>
                    <div>
                      <span className="text-[11px] text-slate-500 block">Write Speed:</span>
                      <span className="text-slate-200">{writeMB} MB/s</span>
                    </div>
                    <div>
                      <span className="text-[11px] text-slate-500 block">Total Read:</span>
                      <span className="text-slate-200">{formatBytes(m.bytes_read)}</span>
                    </div>
                    <div>
                      <span className="text-[11px] text-slate-500 block">RAM Overhead:</span>
                      <span className="text-slate-200">{formatBytes(m.memory_usage)}</span>
                    </div>
                  </div>

                  {/* Source Parts Chips */}
                  {m.source_parts && m.source_parts.length > 0 && (
                    <div className="pt-1">
                      <span className="text-[10px] text-slate-500 block mb-1">Source Parts:</span>
                      <div className="flex flex-wrap gap-1">
                        {m.source_parts.slice(0, 8).map((sp) => (
                          <span
                            key={sp}
                            className="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-[10px] font-mono text-slate-400"
                          >
                            {sp}
                          </span>
                        ))}
                        {m.source_parts.length > 8 && (
                          <span className="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-[10px] font-mono text-slate-500">
                            +{m.source_parts.length - 8} more
                          </span>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Asynchronous Mutations Monitor */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-sky-500/10 text-sky-400 border border-sky-500/20">
              <Activity className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white">Table Mutations (system.mutations)</h3>
              <p className="text-xs text-slate-400">
                Track asynchronous <code className="text-sky-400">ALTER TABLE DELETE/UPDATE</code> execution
              </p>
            </div>
          </div>
        </div>

        {mutations.length === 0 ? (
          <div className="p-8 text-center bg-slate-950/40 rounded-xl border border-slate-800/80">
            <p className="text-xs text-slate-500">No recorded table mutations found.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-950/80 text-slate-400 border-b border-slate-800 font-mono">
                <tr>
                  <th className="py-2.5 px-3">Mutation ID</th>
                  <th className="py-2.5 px-3">Command</th>
                  <th className="py-2.5 px-3">Created</th>
                  <th className="py-2.5 px-3">Parts To Do</th>
                  <th className="py-2.5 px-3 text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 font-mono">
                {mutations.map((mut) => (
                  <tr key={mut.mutation_id} className="hover:bg-slate-800/30">
                    <td className="py-2.5 px-3 font-semibold text-slate-300">{mut.mutation_id}</td>
                    <td className="py-2.5 px-3 text-slate-300 max-w-xs truncate" title={mut.command}>
                      {mut.command}
                    </td>
                    <td className="py-2.5 px-3 text-slate-500">
                      {new Date(mut.create_time).toLocaleTimeString()}
                    </td>
                    <td className="py-2.5 px-3 text-slate-300">{mut.parts_to_do}</td>
                    <td className="py-2.5 px-3 text-right">
                      {mut.is_done ? (
                        <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-950 text-emerald-300 border border-emerald-800 font-bold">
                          COMPLETED
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded text-[10px] bg-amber-950 text-amber-300 border border-amber-800 font-bold animate-pulse">
                          RUNNING
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
