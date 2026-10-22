import React, { useState, useEffect } from 'react';
import type { TableColumnsResponse } from '../types';
import { Check, Copy, Database, HardDrive, Sparkles, TrendingDown, Zap } from 'lucide-react';

interface CodecOptimizerProps {
  database: string;
  table: string;
}

export const CodecOptimizer: React.FC<CodecOptimizerProps> = ({ database, table }) => {
  const [data, setData] = useState<TableColumnsResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [copiedCol, setCopiedCol] = useState<string | null>(null);

  useEffect(() => {
    async function fetchAudit() {
      try {
        setLoading(true);
        const res = await fetch(`/api/columns?database=${database}&table=${table}`);
        if (res.ok) {
          const resData: TableColumnsResponse = await res.json();
          setData(resData);
        }
      } catch {
        // Ignored
      } finally {
        setLoading(false);
      }
    }
    fetchAudit();
  }, [database, table]);

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const handleCopySQL = (colName: string, sql: string) => {
    navigator.clipboard.writeText(sql);
    setCopiedCol(colName);
    setTimeout(() => setCopiedCol(null), 2000);
  };

  if (loading && !data) {
    return (
      <div className="flex items-center justify-center p-16">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-amber-500 border-t-transparent rounded-full animate-spin" />
          <p className="text-xs text-slate-400 font-mono">Auditing column codecs & compression ratios...</p>
        </div>
      </div>
    );
  }

  if (!data || data.columns.length === 0) {
    return (
      <div className="p-12 text-center bg-slate-900/40 rounded-2xl border border-slate-800">
        <Database className="w-10 h-10 text-slate-600 mx-auto mb-3" />
        <h3 className="text-sm font-semibold text-slate-300">No Columns Audited</h3>
        <p className="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
          No column storage metrics found for {database}.{table}.
        </p>
      </div>
    );
  }

  const overallSavingsPct = data.total_compressed_bytes > 0
    ? ((data.potential_savings_bytes / data.total_compressed_bytes) * 100).toFixed(1)
    : '0.0';

  return (
    <div className="space-y-6">
      {/* High-Level Compression Savings Cockpit */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Total Compressed on Disk</span>
            <HardDrive className="w-3.5 h-3.5 text-sky-400" />
          </div>
          <div className="text-xl font-bold text-white font-mono tracking-tight">
            {formatBytes(data.total_compressed_bytes)}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Uncompressed: {formatBytes(data.total_uncompressed_bytes)}
          </div>
        </div>

        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Potential Disk Savings</span>
            <TrendingDown className="w-3.5 h-3.5 text-emerald-400" />
          </div>
          <div className="text-xl font-bold text-emerald-400 font-mono tracking-tight">
            {formatBytes(data.potential_savings_bytes)}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            {overallSavingsPct}% total disk reduction
          </div>
        </div>

        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Columns Analyzed</span>
            <Sparkles className="w-3.5 h-3.5 text-amber-400" />
          </div>
          <div className="text-xl font-bold text-white font-mono tracking-tight">
            {data.total_columns} columns
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            {data.columns.filter((c) => c.savings_estimate_percent > 0).length} optimization opportunities
          </div>
        </div>
      </div>

      {/* Column Details & Codec Recommendations Table */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 overflow-hidden shadow-sm">
        <div className="p-4 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="w-4 h-4 text-amber-400" />
            <h3 className="text-sm font-semibold text-white">Column Compression & Codec Recommendations</h3>
          </div>
          <span className="text-xs text-slate-500 font-mono">
            {database}.{table}
          </span>
        </div>

        <div className="divide-y divide-slate-800/80">
          {data.columns.map((col) => {
            const hasOptimization = col.savings_estimate_percent > 0;
            return (
              <div key={col.column} className="p-4 hover:bg-slate-800/20 transition-colors space-y-2.5">
                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
                  <div className="flex items-center gap-2.5">
                    <span className="text-xs font-mono font-bold text-white">{col.column}</span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-950 text-slate-400 border border-slate-800">
                      {col.type}
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-900 text-slate-400">
                      Codec: {col.compression_codec || 'Default(LZ4)'}
                    </span>
                  </div>

                  <div className="flex items-center gap-4 text-xs font-mono">
                    <span className="text-slate-400">
                      Disk: <strong className="text-slate-200">{formatBytes(col.data_compressed_bytes)}</strong>
                    </span>
                    <span className="text-slate-400">
                      Ratio: <strong className="text-sky-400">{col.compression_ratio.toFixed(2)}x</strong>
                    </span>
                    {hasOptimization ? (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-950 text-emerald-300 border border-emerald-800 font-bold">
                        -{col.savings_estimate_percent.toFixed(0)}% Savings
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800 text-slate-400 border border-slate-700">
                        Optimal
                      </span>
                    )}
                  </div>
                </div>

                {/* Rationale & Recommendation Action */}
                {hasOptimization && (
                  <div className="p-3 bg-slate-950/70 rounded-xl border border-slate-800/90 text-xs space-y-2">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
                      <div className="text-slate-300">
                        <strong className="text-amber-300">Recommended: </strong>
                        <code className="text-emerald-400 font-mono bg-slate-900 px-1.5 py-0.5 rounded">
                          {col.recommended_codec}
                        </code>
                        <span className="text-slate-400 ml-2">
                          (Est. savings: ~{formatBytes(col.projected_savings_bytes)})
                        </span>
                      </div>

                      <button
                        onClick={() => handleCopySQL(col.column, col.alter_sql)}
                        className="flex items-center gap-1.5 px-3 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px] font-medium transition-colors shrink-0 cursor-pointer"
                      >
                        {copiedCol === col.column ? (
                          <Check className="w-3.5 h-3.5 text-emerald-400" />
                        ) : (
                          <Copy className="w-3.5 h-3.5" />
                        )}
                        <span>{copiedCol === col.column ? 'Copied SQL' : 'Copy ALTER SQL'}</span>
                      </button>
                    </div>

                    <p className="text-[11px] text-slate-400">{col.rationale}</p>

                    <div className="p-2 rounded bg-slate-900 border border-slate-800 font-mono text-[11px] text-slate-300 overflow-x-auto">
                      {col.alter_sql}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
