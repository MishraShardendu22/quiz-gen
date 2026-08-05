"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Session } from "@/src/lib/api";

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadSessions() {
      try {
        const data = await api.getSessions();
        setSessions(data || []);
      } catch (err: any) {
        setError(err.message || "Failed to load sessions");
      } finally {
        setLoading(false);
      }
    }
    loadSessions();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Quiz Generation Sessions</h1>
        <p className="text-gray-600 text-sm">View all generated quiz sessions and their status</p>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-6 text-gray-500 text-sm">Loading sessions...</div>
        ) : sessions.length === 0 ? (
          <div className="p-6 text-gray-500 text-sm">No sessions found.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-gray-600">
              <thead className="bg-gray-50 text-gray-700 font-medium border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3">Session ID</th>
                  <th className="px-6 py-3">Topic ID</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3">Requested Count</th>
                  <th className="px-6 py-3">Generated Count</th>
                  <th className="px-6 py-3">Created At</th>
                  <th className="px-6 py-3 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {sessions.map((session) => (
                  <tr key={session.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-mono text-xs text-gray-900">
                      {session.id}
                    </td>
                    <td className="px-6 py-4 font-mono text-xs text-gray-600">
                      {session.topic_id}
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={session.status} />
                    </td>
                    <td className="px-6 py-4">{session.requested_count}</td>
                    <td className="px-6 py-4">{session.generated_count}</td>
                    <td className="px-6 py-4 text-xs">
                      {new Date(session.created_at * 1000).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        href={`/sessions/${session.id}`}
                        className="px-3 py-1.5 border border-gray-300 rounded text-gray-700 hover:bg-gray-100 text-xs font-medium"
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
