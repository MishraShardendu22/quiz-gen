"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, UsageReport } from "@/src/lib/api";

export default function UsagePage() {
  const [report, setReport] = useState<UsageReport | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadUsage() {
      try {
        const data = await api.getUsage();
        setReport(data);
      } catch (err: any) {
        setError(err.message || "Failed to load usage report");
      } finally {
        setLoading(false);
      }
    }
    loadUsage();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Token Usage & Costs</h1>
        <p className="text-slate-400 text-sm mt-1">Aggregated LLM token consumption and estimated costs</p>
      </div>

      {error && (
        <div className="p-4 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-sm">
          {error}
        </div>
      )}

      {/* Metrics Header */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total Tokens</p>
          <p className="text-3xl font-black text-white mt-2">
            {loading ? "-" : report?.total_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-blue-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-blue-400 uppercase tracking-wider">Prompt Tokens</p>
          <p className="text-3xl font-black text-blue-300 mt-2">
            {loading ? "-" : report?.total_prompt_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-purple-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-purple-400 uppercase tracking-wider">Completion Tokens</p>
          <p className="text-3xl font-black text-purple-300 mt-2">
            {loading ? "-" : report?.total_completion_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-emerald-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-emerald-400 uppercase tracking-wider">Estimated Cost</p>
          <p className="text-3xl font-black text-emerald-300 mt-2">
            {loading ? "-" : `$${report?.estimated_cost.toFixed(4)}`}
          </p>
        </div>
      </div>

      {/* Session Breakdown Table */}
      <div className="bg-slate-900/90 rounded-xl border border-slate-800 shadow-xl overflow-hidden backdrop-blur-md">
        <div className="p-6 border-b border-slate-800/80">
          <h2 className="text-lg font-bold text-white">Per-Session Usage Breakdown</h2>
        </div>

        {loading ? (
          <div className="p-6 text-slate-400 text-sm">Loading usage report...</div>
        ) : !report || report.sessions.length === 0 ? (
          <div className="p-6 text-slate-400 text-sm">No usage records found.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300 min-w-[700px]">
              <thead className="bg-slate-800/60 text-slate-400 font-semibold tracking-wider uppercase text-xs border-b border-slate-800">
                <tr>
                  <th className="px-6 py-3.5">Session ID</th>
                  <th className="px-6 py-3.5">Prompt Tokens</th>
                  <th className="px-6 py-3.5">Completion Tokens</th>
                  <th className="px-6 py-3.5">Total Tokens</th>
                  <th className="px-6 py-3.5">Estimated Cost</th>
                  <th className="px-6 py-3.5 text-right whitespace-nowrap">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {report.sessions.map((item) => (
                  <tr key={item.session_id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="px-6 py-4 font-mono text-xs text-white whitespace-nowrap">
                      {item.session_id}
                    </td>
                    <td className="px-6 py-4 font-medium text-slate-300 whitespace-nowrap">{item.prompt_tokens.toLocaleString()}</td>
                    <td className="px-6 py-4 font-medium text-slate-300 whitespace-nowrap">{item.completion_tokens.toLocaleString()}</td>
                    <td className="px-6 py-4 font-bold text-white whitespace-nowrap">
                      {item.total_tokens.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 font-mono text-xs text-emerald-400 font-bold whitespace-nowrap">
                      ${item.estimated_cost.toFixed(4)}
                    </td>
                    <td className="px-6 py-4 text-right whitespace-nowrap">
                      <Link
                        href={`/sessions/${item.session_id}`}
                        className="px-3.5 py-1.5 bg-indigo-600/20 text-indigo-300 hover:bg-indigo-600 hover:text-white border border-indigo-500/30 rounded-lg text-xs font-semibold whitespace-nowrap transition-all inline-block"
                      >
                        View Details
                      </Link>
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
}
