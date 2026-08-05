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
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 text-sm">Overview of quiz generation sessions and stats</p>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-gray-500 uppercase tracking-wider">Total Sessions</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{loading ? "-" : totalSessions}</p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-amber-600 uppercase tracking-wider">Pending</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{loading ? "-" : pendingCount}</p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-blue-600 uppercase tracking-wider">Processing</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{loading ? "-" : processingCount}</p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-green-600 uppercase tracking-wider">Completed</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{loading ? "-" : completedCount}</p>
        </div>

        <div className="bg-white p-5 rounded-lg border border-gray-200 shadow-sm">
          <p className="text-xs font-medium text-red-600 uppercase tracking-wider">Failed</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{loading ? "-" : failedCount}</p>
        </div>
      </div>

      {/* Quick Links */}
      <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Links</h2>
        <div className="flex flex-wrap gap-4">
          <Link
            href="/topics"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium"
          >
            Browse Topics & Generate Questions
          </Link>
          <Link
            href="/sessions"
            className="px-4 py-2 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 text-sm font-medium"
          >
            View All Sessions
          </Link>
        </div>
      </div>

      {/* Recent Sessions */}
      <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
        <div className="p-6 border-b border-gray-200 flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900">Recent Sessions</h2>
          <Link href="/sessions" className="text-sm font-medium text-blue-600 hover:text-blue-800">
            View All →
          </Link>
        </div>

        {loading ? (
          <div className="p-6 text-gray-500 text-sm">Loading sessions...</div>
        ) : recentSessions.length === 0 ? (
          <div className="p-6 text-gray-500 text-sm">No sessions generated yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-gray-600">
              <thead className="bg-gray-50 text-gray-700 font-medium border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3">Session ID</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3">Requested</th>
                  <th className="px-6 py-3">Generated</th>
                  <th className="px-6 py-3">Created At</th>
                  <th className="px-6 py-3">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {recentSessions.map((session) => (
                  <tr key={session.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-mono text-xs text-gray-900">
                      {session.id}
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={session.status} />
                    </td>
                    <td className="px-6 py-4">{session.requested_count}</td>
                    <td className="px-6 py-4">{session.generated_count}</td>
                    <td className="px-6 py-4 text-xs">
                      {new Date(session.created_at * 1000).toLocaleString()}
                    </td>
                    <td className="px-6 py-4">
                      <Link
                        href={`/sessions/${session.id}`}
                        className="text-blue-600 hover:text-blue-800 font-medium"
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
  let colorClass = "bg-gray-100 text-gray-800";
  if (status === "pending") colorClass = "bg-amber-100 text-amber-800";
  if (status === "processing") colorClass = "bg-blue-100 text-blue-800";
  if (status === "completed") colorClass = "bg-green-100 text-green-800";
  if (status === "failed") colorClass = "bg-red-100 text-red-800";

  return (
    <span className={`inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold ${colorClass}`}>
      {status}
    </span>
  );
}
