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
        <h1 className="text-2xl font-bold text-gray-900">Token Usage & Costs</h1>
        <p className="text-gray-600 text-sm">Aggregated LLM token consumption and estimated costs</p>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}

      {/* Metrics Header */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wider">Total Tokens</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">
            {loading ? "-" : report?.total_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wider">Prompt Tokens</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">
            {loading ? "-" : report?.total_prompt_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wider">Completion Tokens</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">
            {loading ? "-" : report?.total_completion_tokens.toLocaleString()}
          </p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wider">Estimated Cost</p>
          <p className="text-2xl font-bold text-green-700 mt-2">
            {loading ? "-" : `$${report?.estimated_cost.toFixed(4)}`}
          </p>
        </div>
      </div>

      {/* Session Breakdown Table */}
      <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
        <div className="p-6 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Per-Session Usage Breakdown</h2>
        </div>

        {loading ? (
          <div className="p-6 text-gray-500 text-sm">Loading usage report...</div>
        ) : !report || report.sessions.length === 0 ? (
          <div className="p-6 text-gray-500 text-sm">No usage records found.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-gray-600">
              <thead className="bg-gray-50 text-gray-700 font-medium border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3">Session ID</th>
                  <th className="px-6 py-3">Prompt Tokens</th>
                  <th className="px-6 py-3">Completion Tokens</th>
                  <th className="px-6 py-3">Total Tokens</th>
                  <th className="px-6 py-3">Estimated Cost</th>
                  <th className="px-6 py-3 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {report.sessions.map((item) => (
                  <tr key={item.session_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-mono text-xs text-gray-900">
                      {item.session_id}
                    </td>
                    <td className="px-6 py-4">{item.prompt_tokens.toLocaleString()}</td>
                    <td className="px-6 py-4">{item.completion_tokens.toLocaleString()}</td>
                    <td className="px-6 py-4 font-semibold text-gray-900">
                      {item.total_tokens.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 font-mono text-xs text-green-700">
                      ${item.estimated_cost.toFixed(4)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        href={`/sessions/${item.session_id}`}
                        className="text-blue-600 hover:text-blue-800 text-xs font-medium"
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
