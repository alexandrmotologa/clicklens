import React, { useState } from 'react';
import type { PartInfo, TablePartsResponse } from '../types';
import { AlertTriangle, Copy, Check, HardDrive, Info, Layers, X } from 'lucide-react';

interface PartsHeatmapProps {
  data: TablePartsResponse | null;
  loading: boolean;
}

export const PartsHeatmap: React.FC<PartsHeatmapProps> = ({ data, loading }) => {
  const [selectedPartition, setSelectedPartition] = useState<string>('');
  const [selectedPart, setSelectedPart] = useState<PartInfo | null>(null);
  const [copiedSQL, setCopiedSQL] = useState<boolean>(false);

  if (loading && !data) {
    return (
      <div className="flex items-center justify-center p-16">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-sky-500 border-t-transparent rounded-full animate-spin" />
          <p className="text-xs text-slate-400 font-mono">Scanning MergeTree parts layout...</p>
        </div>
      </div>
    );
  }

  if (!data || data.partitions.length === 0) {
    return (
      <div className="p-12 text-center bg-slate-900/40 rounded-2xl border border-slate-800">
        <Layers className="w-10 h-10 text-slate-600 mx-auto mb-3" />
        <h3 className="text-sm font-semibold text-slate-300">No Parts Found</h3>
        <p className="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
          The selected table has no recorded parts in system.parts, or is not using a MergeTree family engine.
        </p>
      </div>
    );
  }

  const activePartitionKey = selectedPartition || data.partitions[0]?.partition;
  const currentPartition = data.partitions.find((p) => p.partition === activePartitionKey) || data.partitions[0];

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const handleCopySQL = (sql: string) => {
    navigator.clipboard.writeText(sql);
    setCopiedSQL(true);
    setTimeout(() => setCopiedSQL(false), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Table Summary Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Table Engine</span>
            <Layers className="w-3.5 h-3.5 text-sky-400" />
          </div>
          <div className="text-xl font-bold text-white tracking-tight">{data.engine}</div>
          <div className="text-[11px] text-slate-500 mt-1">
            {data.database}.{data.table}
          </div>
        </div>

        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Active vs Inactive Parts</span>
            <span
              className={`text-[10px] px-1.5 py-0.5 rounded font-bold ${
                data.health_rating === 'HEALTHY'
                  ? 'bg-emerald-950 text-emerald-300 border border-emerald-800'
                  : data.health_rating === 'DEGRADED'
                  ? 'bg-amber-950 text-amber-300 border border-amber-800'
                  : 'bg-rose-950 text-rose-300 border border-rose-800'
              }`}
            >
              {data.health_rating}
            </span>
          </div>
          <div className="text-xl font-bold text-white tracking-tight">
            {data.total_active_parts} <span className="text-xs font-normal text-slate-500">active</span> /{' '}
            <span className="text-slate-400">{data.total_inactive_parts}</span>{' '}
            <span className="text-xs font-normal text-slate-500">pending drop</span>
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Max in partition: {data.max_part_count} (Limit: 300)
          </div>
        </div>

        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Disk Allocation</span>
            <HardDrive className="w-3.5 h-3.5 text-amber-400" />
          </div>
          <div className="text-xl font-bold text-white tracking-tight">
            {formatBytes(data.total_bytes_on_disk)}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Uncompressed: {formatBytes(data.total_uncompressed)}
          </div>
        </div>

        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
            <span>Overall Compression</span>
            <span className="text-xs font-mono font-semibold text-sky-400">
              {data.overall_ratio.toFixed(2)}x
            </span>
          </div>
          <div className="text-xl font-bold text-emerald-400 tracking-tight">
            {(((data.total_uncompressed - data.total_bytes_on_disk) / (data.total_uncompressed || 1)) * 100).toFixed(1)}%
          </div>
          <div className="text-[11px] text-slate-500 mt-1">Storage space saved</div>
        </div>
      </div>

      {/* Micro-Batching Ingestion Anomaly Banner */}
      {data.micro_batching_warning && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
          <div className="text-xs">
            <span className="font-semibold text-amber-200">Ingestion Micro-Batching Anomaly Detected: </span>
            <span className="text-amber-300/90 leading-relaxed block mt-0.5">
              {data.micro_batching_details || 'High frequency inserts are generating too many small parts in short intervals. Consider buffering inserts or batching streaming writes to prevent MergeTree write throttling.'}
            </span>
          </div>
        </div>
      )}

      {/* Part Age Distribution Histogram */}
      {data.part_age_buckets && data.part_age_buckets.length > 0 && (
        <div className="p-4 bg-slate-900/80 rounded-xl border border-slate-800 space-y-3">
          <div className="flex items-center justify-between text-xs">
            <span className="font-semibold text-white">Part Age Histogram & Ingestion Profile</span>
            <span className="text-[11px] text-slate-400">Time since part creation or merge</span>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 pt-1">
            {data.part_age_buckets.map((bucket) => {
              const maxCount = Math.max(...(data.part_age_buckets?.map((b) => b.part_count) || [1]));
              const pct = maxCount > 0 ? (bucket.part_count / maxCount) * 100 : 0;
              const isRecent = bucket.label.includes('5 min') || bucket.label.includes('30 min');

              return (
                <div key={bucket.label} className="p-3 bg-slate-950/70 rounded-lg border border-slate-800/80">
                  <div className="flex items-center justify-between text-[11px] text-slate-400 mb-1">
                    <span className="font-medium text-slate-300">{bucket.label}</span>
                    <span className={`font-mono font-bold ${isRecent && bucket.part_count > 10 ? 'text-amber-400' : 'text-sky-400'}`}>
                      {bucket.part_count} parts
                    </span>
                  </div>

                  <div className="h-1.5 w-full bg-slate-900 rounded-full overflow-hidden mb-2">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        isRecent && bucket.part_count > 10 ? 'bg-amber-400' : 'bg-sky-500'
                      }`}
                      style={{ width: `${Math.max(4, pct)}%` }}
                    />
                  </div>

                  <div className="text-[10px] text-slate-500 flex justify-between font-mono">
                    <span>{bucket.total_rows.toLocaleString()} rows</span>
                    <span>{formatBytes(bucket.total_bytes)}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Partition Health Warning Banner (if needed) */}
      {data.max_part_count > 100 && (
        <div className="p-3.5 rounded-xl bg-amber-950/40 border border-amber-800/60 flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
          <div className="text-xs">
            <span className="font-semibold text-amber-200">High Part Count Warning: </span>
            <span className="text-amber-300/90">
              Partition <code className="bg-amber-950/80 px-1 rounded">{data.max_parts_partition}</code> has{' '}
              {data.max_part_count} active parts. ClickHouse default limit is 300 active parts before write
              throttling kicks in. Consider consolidating with background merges or executing{' '}
              <code className="bg-amber-950/80 px-1 rounded font-mono">OPTIMIZE TABLE FINAL</code>.
            </span>
          </div>
        </div>
      )}

      {/* Partition Navigation Tabs */}
      <div className="flex items-center gap-2 overflow-x-auto pb-2 border-b border-slate-800">
        <span className="text-xs font-medium text-slate-500 mr-1 shrink-0">Partitions:</span>
        {data.partitions.map((part) => (
          <button
            key={part.partition}
            onClick={() => setSelectedPartition(part.partition)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-all shrink-0 cursor-pointer ${
              activePartitionKey === part.partition
                ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                : 'bg-slate-900/60 text-slate-400 hover:text-slate-200 border border-slate-800'
            }`}
          >
            <span>{part.partition}</span>
            <span
              className={`w-1.5 h-1.5 rounded-full ${
                part.health_status === 'HEALTHY'
                  ? 'bg-emerald-400'
                  : part.health_status === 'WARNING'
                  ? 'bg-amber-400'
                  : 'bg-rose-400'
              }`}
            />
            <span className="text-[11px] font-mono text-slate-500">({part.active_parts})</span>
          </button>
        ))}
      </div>

      {/* Selected Partition Detailed Cockpit */}
      {currentPartition && (
        <div className="bg-slate-900/70 rounded-2xl border border-slate-800 p-5 space-y-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-white">
                  Partition: <span className="font-mono text-sky-400">{currentPartition.partition}</span>
                </h3>
                <span
                  className={`text-[10px] px-2 py-0.5 rounded font-bold uppercase tracking-wider ${
                    currentPartition.health_status === 'HEALTHY'
                      ? 'bg-emerald-950 text-emerald-300 border border-emerald-800'
                      : currentPartition.health_status === 'WARNING'
                      ? 'bg-amber-950 text-amber-300 border border-amber-800'
                      : 'bg-rose-950 text-rose-300 border border-rose-800'
                  }`}
                >
                  {currentPartition.health_status}
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                {currentPartition.total_rows.toLocaleString()} total rows across{' '}
                {currentPartition.active_parts} active parts ({formatBytes(currentPartition.total_bytes_on_disk)})
              </p>
            </div>

            {/* Actions: Optimize Button */}
            <div className="flex items-center gap-2">
              <button
                onClick={() => handleCopySQL(currentPartition.optimization_sql)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-300 border border-slate-700 transition-colors cursor-pointer"
                title="Copy OPTIMIZE TABLE statement"
              >
                {copiedSQL ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copiedSQL ? 'Copied SQL' : 'Copy Optimize SQL'}</span>
              </button>
            </div>
          </div>

          {/* Fragmentation Index Meter */}
          <div>
            <div className="flex justify-between text-xs text-slate-400 mb-1.5">
              <span>Fragmentation Index</span>
              <span className="font-mono font-medium text-slate-300">
                {currentPartition.fragmentation_score.toFixed(1)}%
              </span>
            </div>
            <div className="h-2 w-full bg-slate-950 rounded-full overflow-hidden border border-slate-800">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  currentPartition.fragmentation_score < 30
                    ? 'bg-emerald-500'
                    : currentPartition.fragmentation_score < 60
                    ? 'bg-amber-500'
                    : 'bg-rose-500'
                }`}
                style={{ width: `${Math.max(5, currentPartition.fragmentation_score)}%` }}
              />
            </div>
          </div>

          {/* Interactive Parts Heatmap Grid */}
          <div>
            <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
              <span className="font-medium text-slate-300">Data Parts Grid</span>
              <div className="flex items-center gap-3 text-[11px]">
                <div className="flex items-center gap-1.5">
                  <div className="w-3 h-3 rounded bg-sky-500/80 border border-sky-400" />
                  <span>Active Part</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <div className="w-3 h-3 rounded bg-slate-800 border border-dashed border-slate-600" />
                  <span>Inactive (Cleanup)</span>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 gap-2 p-3 bg-slate-950/70 rounded-xl border border-slate-800/80 max-h-96 overflow-y-auto">
              {currentPartition.parts?.map((p) => {
                const isSelected = selectedPart?.name === p.name;
                return (
                  <button
                    key={p.name}
                    onClick={() => setSelectedPart(p)}
                    className={`p-2 rounded-lg text-left transition-all border relative group cursor-pointer ${
                      isSelected
                        ? 'ring-2 ring-sky-400 border-sky-400 bg-sky-950/50'
                        : p.active
                        ? 'bg-sky-950/30 hover:bg-sky-900/40 border-sky-800/50 text-slate-200'
                        : 'bg-slate-900/40 hover:bg-slate-800/50 border-slate-800 border-dashed text-slate-500'
                    }`}
                  >
                    <div className="text-[10px] font-mono font-medium truncate text-sky-300" title={p.name}>
                      {p.name.replace(`${currentPartition.partition}_`, '')}
                    </div>
                    <div className="text-[11px] font-bold text-white mt-1">
                      {formatBytes(p.bytes_on_disk)}
                    </div>
                    <div className="text-[9px] text-slate-400 mt-0.5">
                      {p.rows > 1000 ? `${(p.rows / 1000).toFixed(0)}k rows` : `${p.rows} rows`}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Part Detail Inspector Drawer / Modal */}
      {selectedPart && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-sky-500/10 text-sky-400 border border-sky-500/20">
                  <Info className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white">MergeTree Part Inspector</h3>
                  <p className="text-xs font-mono text-slate-400 truncate max-w-xs">{selectedPart.name}</p>
                </div>
              </div>
              <button
                onClick={() => setSelectedPart(null)}
                className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="grid grid-cols-2 gap-3 text-xs">
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Status:</span>
                <span className="font-semibold text-white">
                  {selectedPart.active ? 'Active (Live Ingestion)' : 'Inactive (Waiting Drop)'}
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Compression Ratio:</span>
                <span className="font-semibold text-sky-400 font-mono">
                  {selectedPart.compression_ratio.toFixed(2)}x
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Bytes On Disk:</span>
                <span className="font-semibold text-white font-mono">
                  {formatBytes(selectedPart.bytes_on_disk)}
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Uncompressed Bytes:</span>
                <span className="font-semibold text-white font-mono">
                  {formatBytes(selectedPart.data_uncompressed_bytes)}
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Rows:</span>
                <span className="font-semibold text-white font-mono">
                  {selectedPart.rows.toLocaleString()}
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Marks Count:</span>
                <span className="font-semibold text-white font-mono">
                  {selectedPart.marks.toLocaleString()} (Index Granularity: 8192)
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Block Range & Level:</span>
                <span className="font-semibold text-white font-mono">
                  Blocks [{selectedPart.min_block_number} .. {selectedPart.max_block_number}], Level {selectedPart.level}
                </span>
              </div>
              <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80">
                <span className="text-slate-500 block">Modification Time:</span>
                <span className="font-semibold text-white font-mono">
                  {new Date(selectedPart.modification_time).toLocaleTimeString()}
                </span>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button
                onClick={() => setSelectedPart(null)}
                className="px-4 py-2 text-xs font-semibold text-slate-950 bg-sky-400 hover:bg-sky-300 rounded-lg transition-colors cursor-pointer"
              >
                Close Inspector
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
