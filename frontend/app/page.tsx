"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Session } from "@/src/lib/api";

export default function DashboardPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadData() {
      try {
        const data = await api.getSessions();
        setSessions(data || []);
      } catch (err: any) {
        setError(err.message || "Failed to load dashboard sessions");
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, []);

  const totalSessions = sessions.length;
  const pendingCount = sessions.filter((s) => s.status === "pending").length;
  const processingCount = sessions.filter((s) => s.status === "processing").length;
  const completedCount = sessions.filter((s) => s.status === "completed").length;
  const failedCount = sessions.filter((s) => s.status === "failed").length;

  const recentSessions = [...sessions].slice(0, 5);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Dashboard Overview</h1>
        <p className="text-slate-400 text-sm mt-1">Real-time statistics for AI quiz generation sessions</p>
      </div>

      {error && (
        <div className="p-4 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-sm">
          {error}
        </div>
      )}

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total Sessions</p>
          <p className="text-3xl font-black text-white mt-2">{loading ? "-" : totalSessions}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-blue-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-blue-400 uppercase tracking-wider">Pending</p>
          <p className="text-3xl font-black text-blue-300 mt-2">{loading ? "-" : pendingCount}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-amber-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-amber-400 uppercase tracking-wider">Processing</p>
          <p className="text-3xl font-black text-amber-300 mt-2">{loading ? "-" : processingCount}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-emerald-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-emerald-400 uppercase tracking-wider">Completed</p>
          <p className="text-3xl font-black text-emerald-300 mt-2">{loading ? "-" : completedCount}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-rose-500/20 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-rose-400 uppercase tracking-wider">Failed</p>
          <p className="text-3xl font-black text-rose-300 mt-2">{loading ? "-" : failedCount}</p>
        </div>
      </div>

      {/* Quick Links Banner */}
      <div className="bg-gradient-to-r from-indigo-900/40 via-slate-900 to-purple-900/40 p-6 rounded-xl border border-indigo-500/20 shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-white">Generate Quiz Questions</h2>
          <p className="text-xs text-slate-300 mt-1">Select a topic from discovered markdown content to begin quiz generation.</p>
        </div>
        <div className="flex flex-wrap gap-3">
          <Link
            href="/topics"
            className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-bold shadow-lg shadow-indigo-600/30 transition-all whitespace-nowrap"
          >
            Browse Topics →
          </Link>
          <Link
            href="/sessions"
            className="px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-xs font-bold border border-slate-700 transition-all whitespace-nowrap"
          >
            View All Sessions
          </Link>
        </div>
      </div>

      {/* Recent Sessions Table */}
      <div className="bg-slate-900/90 rounded-xl border border-slate-800 shadow-xl overflow-hidden backdrop-blur-md">
        <div className="p-6 border-b border-slate-800/80 flex justify-between items-center">
          <h2 className="text-lg font-bold text-white">Recent Sessions</h2>
          <Link href="/sessions" className="text-xs font-semibold text-indigo-400 hover:text-indigo-300">
            View All →
          </Link>
        </div>

        {loading ? (
          <div className="p-6 text-slate-400 text-sm">Loading sessions...</div>
        ) : recentSessions.length === 0 ? (
          <div className="p-6 text-slate-400 text-sm">No sessions generated yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-800/60 text-slate-400 font-semibold tracking-wider uppercase text-xs border-b border-slate-800">
                <tr>
                  <th className="px-6 py-3.5">Session ID</th>
                  <th className="px-6 py-3.5">Status</th>
                  <th className="px-6 py-3.5">Requested</th>
                  <th className="px-6 py-3.5">Generated</th>
                  <th className="px-6 py-3.5">Created At</th>
                  <th className="px-6 py-3.5 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {recentSessions.map((session) => (
                  <tr key={session.id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="px-6 py-4 font-mono text-xs text-white">
                      {session.id}
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={session.status} />
                    </td>
                    <td className="px-6 py-4 font-medium text-white">{session.requested_count}</td>
                    <td className="px-6 py-4 font-medium text-white">{session.generated_count}</td>
                    <td className="px-6 py-4 text-xs text-slate-400 whitespace-nowrap">
                      {new Date(session.created_at * 1000).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-right whitespace-nowrap">
                      <Link
                        href={`/sessions/${session.id}`}
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

function StatusBadge({ status }: { status: string }) {
  let colorClass = "bg-slate-800 text-slate-300 border-slate-700";
  if (status === "pending") colorClass = "bg-blue-500/10 text-blue-400 border-blue-500/30";
  if (status === "processing") colorClass = "bg-amber-500/10 text-amber-400 border-amber-500/30";
  if (status === "completed") colorClass = "bg-emerald-500/10 text-emerald-400 border-emerald-500/30";
  if (status === "failed") colorClass = "bg-rose-500/10 text-rose-400 border-rose-500/30";

  return (
    <span className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-bold border capitalize ${colorClass}`}>
      {status}
    </span>
  );
}
