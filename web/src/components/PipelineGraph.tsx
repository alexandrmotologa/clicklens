import React, { useState } from 'react';
import type { PipelineAnalysis } from '../types';
import { AlertCircle, ArrowDown, Cpu, GitFork, Play, FileText, Check, Copy } from 'lucide-react';

interface PipelineGraphProps {
  onProfileQuery: (query: string, mode: string) => Promise<PipelineAnalysis | null>;
}

export const PipelineGraph: React.FC<PipelineGraphProps> = ({ onProfileQuery }) => {
  const [queryInput, setQueryInput] = useState<string>(
    'SELECT event_type, host, count(), avg(cpu_usage) FROM telemetry.events WHERE event_time >= now() - INTERVAL 1 HOUR GROUP BY event_type, host ORDER BY count() DESC LIMIT 100'
  );
  const [explainMode, setExplainMode] = useState<'PIPELINE' | 'PLAN'>('PIPELINE');
  const [loading, setLoading] = useState<boolean>(false);
  const [analysis, setAnalysis] = useState<PipelineAnalysis | null>(null);
  const [viewTab, setViewTab] = useState<'GRAPH' | 'RAW'>('GRAPH');
  const [copiedRaw, setCopiedRaw] = useState(false);

  const sampleQueries = [
    {
      label: 'Telemetry Group By & Aggregation',
      sql: 'SELECT event_type, host, count(), avg(cpu_usage) FROM telemetry.events WHERE event_time >= now() - INTERVAL 1 HOUR GROUP BY event_type, host ORDER BY count() DESC LIMIT 100',
    },
    {
      label: 'Multi-Partition Scan & Sort',
      sql: 'SELECT * FROM telemetry.events WHERE partition = 202610 ORDER BY event_time DESC LIMIT 50',
    },
    {
      label: 'High-Volume Window Aggregation',
      sql: 'SELECT event_time, host, sum(cpu_usage) OVER (PARTITION BY host ORDER BY event_time ROWS BETWEEN 10 PRECEDING AND CURRENT ROW) FROM telemetry.events',
    },
  ];

  const handleRunProfile = async () => {
    if (!queryInput.trim()) return;
    setLoading(true);
    try {
      const res = await onProfileQuery(queryInput, explainMode);
      if (res) {
        setAnalysis(res);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCopyRaw = () => {
    if (analysis?.raw_text) {
      navigator.clipboard.writeText(analysis.raw_text);
      setCopiedRaw(true);
      setTimeout(() => setCopiedRaw(false), 2000);
    }
  };

  return (
    <div className="space-y-6">
      {/* Query Profiler Input Box */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-800 p-5 space-y-4 shadow-sm">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-sky-500/10 text-sky-400 border border-sky-500/20">
              <Cpu className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white">Pipeline Execution Profiler</h3>
              <p className="text-xs text-slate-400">
                Inspect thread concurrency DAG and memory spill markers via <code className="text-sky-400">EXPLAIN PIPELINE</code>
              </p>
            </div>
          </div>

          {/* Mode Switcher */}
          <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs">
            <button
              onClick={() => setExplainMode('PIPELINE')}
              className={`px-3 py-1 rounded font-medium transition-all cursor-pointer ${
                explainMode === 'PIPELINE'
                  ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              EXPLAIN PIPELINE
            </button>
            <button
              onClick={() => setExplainMode('PLAN')}
              className={`px-3 py-1 rounded font-medium transition-all cursor-pointer ${
                explainMode === 'PLAN'
                  ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              EXPLAIN PLAN
            </button>
          </div>
        </div>

        {/* Quick Query Templates */}
        <div className="flex items-center gap-2 overflow-x-auto pb-1">
          <span className="text-[11px] text-slate-500 shrink-0">Templates:</span>
          {sampleQueries.map((q) => (
            <button
              key={q.label}
              onClick={() => setQueryInput(q.sql)}
              className="text-[11px] px-2.5 py-1 rounded-md bg-slate-950 hover:bg-slate-800 text-slate-400 hover:text-slate-200 border border-slate-800 transition-colors shrink-0 cursor-pointer"
            >
              {q.label}
            </button>
          ))}
        </div>

        {/* Editor Textarea */}
        <div className="relative">
          <textarea
            value={queryInput}
            onChange={(e) => setQueryInput(e.target.value)}
            rows={3}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3.5 text-xs text-slate-200 font-mono focus:outline-none focus:border-sky-500 placeholder-slate-600 resize-none shadow-inner"
            placeholder="Enter SQL analytical query to profile execution pipeline..."
          />
        </div>

        <div className="flex justify-end">
          <button
            onClick={handleRunProfile}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-sky-400 hover:bg-sky-300 text-slate-950 text-xs font-semibold shadow-lg shadow-sky-500/20 transition-all disabled:opacity-50 cursor-pointer"
          >
            {loading ? (
              <div className="w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
            ) : (
              <Play className="w-3.5 h-3.5 fill-current" />
            )}
            <span>Profile Execution Pipeline</span>
          </button>
        </div>
      </div>

      {/* Profiler Results View */}
      {analysis && (
        <div className="space-y-4">
          {/* Concurrency & Bottleneck Metrics Bar */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Max Concurrency</span>
              <div className="text-xl font-bold text-sky-400 font-mono mt-0.5">
                {analysis.max_concurrency} <span className="text-xs font-normal text-slate-500">threads</span>
              </div>
            </div>

            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Identified Bottlenecks</span>
              <div
                className={`text-xl font-bold font-mono mt-0.5 ${
                  analysis.bottlenecks.length > 0 ? 'text-amber-400' : 'text-emerald-400'
                }`}
              >
                {analysis.bottlenecks.length}{' '}
                <span className="text-xs font-normal text-slate-500">
                  {analysis.bottlenecks.length > 0 ? 'collapse warning' : 'optimal parallelism'}
                </span>
              </div>
            </div>

            <div className="p-3.5 bg-slate-900/80 rounded-xl border border-slate-800">
              <span className="text-[11px] text-slate-400 block">Execution Pipeline Steps</span>
              <div className="text-xl font-bold text-white font-mono mt-0.5">
                {analysis.nodes.length} <span className="text-xs font-normal text-slate-500">transforms</span>
              </div>
            </div>
          </div>

          {/* Spillover Warnings if Any */}
          {analysis.spill_alerts && analysis.spill_alerts.length > 0 && (
            <div className="space-y-2">
              {analysis.spill_alerts.map((spill, idx) => (
                <div
                  key={idx}
                  className="p-3.5 rounded-xl bg-amber-950/40 border border-amber-800/60 flex items-start gap-3"
                >
                  <AlertCircle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
                  <div className="text-xs">
                    <span className="font-semibold text-amber-200">Potential Memory Spill Risk: </span>
                    <span className="text-amber-300/90">{spill.advice}</span>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Bottlenecks Detailed Alert */}
          {analysis.bottlenecks.map((b) => (
            <div
              key={b.node_id}
              className="p-3.5 rounded-xl bg-rose-950/30 border border-rose-800/50 flex items-start gap-3"
            >
              <GitFork className="w-5 h-5 text-rose-400 shrink-0 mt-0.5" />
              <div className="text-xs space-y-1">
                <div className="font-semibold text-rose-200">
                  Concurrency Bottleneck on Step: <code className="bg-rose-950/80 px-1 rounded">{b.step_type}</code>
                </div>
                <div className="text-rose-300/90">{b.description}</div>
                <div className="text-slate-400">
                  <strong>Recommendation:</strong> {b.recommendation}
                </div>
              </div>
            </div>
          ))}

          {/* View Tab Toggle (Graph vs Raw) */}
          <div className="flex items-center justify-between border-b border-slate-800 pb-2">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setViewTab('GRAPH')}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer ${
                  viewTab === 'GRAPH'
                    ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <GitFork className="w-3.5 h-3.5" />
                <span>Execution Stream DAG</span>
              </button>
              <button
                onClick={() => setViewTab('RAW')}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer ${
                  viewTab === 'RAW'
                    ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <FileText className="w-3.5 h-3.5" />
                <span>Raw Text Tree</span>
              </button>
            </div>

            {viewTab === 'RAW' && (
              <button
                onClick={handleCopyRaw}
                className="flex items-center gap-1 text-xs text-slate-400 hover:text-white px-2 py-1 rounded bg-slate-900 border border-slate-800 cursor-pointer"
              >
                {copiedRaw ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copiedRaw ? 'Copied' : 'Copy Output'}</span>
              </button>
            )}
          </div>

          {/* Graphical Stream DAG */}
          {viewTab === 'GRAPH' ? (
            <div className="p-6 bg-slate-900/60 rounded-2xl border border-slate-800 space-y-4">
              <div className="flex flex-col items-center max-w-xl mx-auto space-y-3">
                {analysis.nodes.map((node, idx) => {
                  const isLast = idx === analysis.nodes.length - 1;
                  return (
                    <React.Fragment key={node.id}>
                      <div
                        className={`w-full p-4 rounded-xl border transition-all ${
                          node.is_bottleneck
                            ? 'bg-rose-950/40 border-rose-700/80 shadow-lg shadow-rose-950/50'
                            : 'bg-slate-950 border-slate-800 hover:border-slate-700'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-1">
                          <div className="flex items-center gap-2">
                            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-900 text-slate-400 border border-slate-800">
                              {node.id}
                            </span>
                            <span className="text-xs font-bold text-white font-mono">{node.step_type}</span>
                          </div>

                          <div className="flex items-center gap-1.5">
                            <span
                              className={`text-[10px] font-mono px-2 py-0.5 rounded-full font-bold ${
                                node.stream_count > 1
                                  ? 'bg-sky-950 text-sky-300 border border-sky-800'
                                  : 'bg-slate-900 text-slate-400 border border-slate-800'
                              }`}
                            >
                              {node.stream_count} {node.stream_count === 1 ? 'stream' : 'parallel streams'}
                            </span>
                          </div>
                        </div>

                        <div className="text-xs text-slate-300 font-mono break-all">{node.label}</div>

                        {node.details && (
                          <div className="text-[11px] text-slate-500 mt-1 font-mono">{node.details}</div>
                        )}
                      </div>

                      {!isLast && (
                        <div className="flex flex-col items-center my-0.5 text-slate-600">
                          <ArrowDown className="w-4 h-4 text-sky-400 animate-bounce" />
                        </div>
                      )}
                    </React.Fragment>
                  );
                })}
              </div>
            </div>
          ) : (
            <div className="p-4 bg-slate-950 rounded-2xl border border-slate-800 font-mono text-xs text-slate-300 overflow-x-auto whitespace-pre leading-relaxed shadow-inner">
              {analysis.raw_text}
            </div>
          )}
        </div>
      )}
    </div>
  );
};
