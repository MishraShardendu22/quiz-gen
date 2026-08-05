"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Session } from "@/src/lib/api";
import { getRetryMap } from "@/src/lib/retryTracker";

interface SessionNode {
  session: Session;
  children: SessionNode[];
  isRetry: boolean;
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionTree, setSessionTree] = useState<SessionNode[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadSessions() {
      try {
        const data = await api.getSessions();
        const rawSessions = data || [];
        setSessions(rawSessions);

        const retryMap = getRetryMap();
        const sessionById = new Map<string, Session>();
        rawSessions.forEach((s) => sessionById.set(s.id, s));

        const parentOf = new Map<string, string>();

        // 1. Explicit retry map
        Object.entries(retryMap).forEach(([childId, parentId]) => {
          if (sessionById.has(childId) && sessionById.has(parentId)) {
            parentOf.set(childId, parentId);
          }
        });

        // 2. Infer retries for sessions on same topic created after a failed session if not in explicit map
        const topicSessionsMap = new Map<string, Session[]>();
        rawSessions.forEach((s) => {
          const list = topicSessionsMap.get(s.topic_id) || [];
          list.push(s);
          topicSessionsMap.set(s.topic_id, list);
        });

        topicSessionsMap.forEach((tSessions) => {
          const sorted = [...tSessions].sort((a, b) => a.created_at - b.created_at);
          for (let i = 0; i < sorted.length; i++) {
            if (sorted[i].status === "failed") {
              for (let j = i + 1; j < sorted.length; j++) {
                const nextS = sorted[j];
                if (!parentOf.has(nextS.id)) {
                  parentOf.set(nextS.id, sorted[i].id);
                  break;
                }
              }
            }
          }
        });

        // Construct tree nodes
        const childrenOfMap = new Map<string, Session[]>();
        const rootSessions: Session[] = [];

        rawSessions.forEach((s) => {
          const parentId = parentOf.get(s.id);
          if (parentId && sessionById.has(parentId)) {
            const list = childrenOfMap.get(parentId) || [];
            list.push(s);
            childrenOfMap.set(parentId, list);
          } else {
            rootSessions.push(s);
          }
        });

        function buildNode(s: Session, isRetry: boolean): SessionNode {
          const children = (childrenOfMap.get(s.id) || [])
            .sort((a, b) => a.created_at - b.created_at)
            .map((child) => buildNode(child, true));

          return { session: s, children, isRetry };
        }

        const tree = rootSessions.map((s) => buildNode(s, false));
        setSessionTree(tree);
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
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Quiz Generation Sessions</h1>
        <p className="text-slate-400 text-sm mt-1">View all quiz sessions and retries in hierarchical order</p>
      </div>

      {error && (
        <div className="p-4 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-sm">
          {error}
        </div>
      )}

      <div className="bg-slate-900/90 rounded-xl border border-slate-800 shadow-xl overflow-hidden backdrop-blur-md">
        {loading ? (
          <div className="p-6 text-slate-400 text-sm">Loading sessions...</div>
        ) : sessionTree.length === 0 ? (
          <div className="p-6 text-slate-400 text-sm">No sessions found.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300 min-w-[700px]">
              <thead className="bg-slate-800/60 text-slate-400 font-semibold tracking-wider uppercase text-xs border-b border-slate-800">
                <tr>
                  <th className="px-6 py-3.5">Session ID</th>
                  <th className="px-6 py-3.5">Topic ID</th>
                  <th className="px-6 py-3.5">Status</th>
                  <th className="px-6 py-3.5">Requested</th>
                  <th className="px-6 py-3.5">Generated</th>
                  <th className="px-6 py-3.5">Tokens Used</th>
                  <th className="px-6 py-3.5">Created At</th>
                  <th className="px-6 py-3.5 text-right whitespace-nowrap">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {sessionTree.map((node) => (
                  <SessionRowTree key={node.session.id} node={node} level={0} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function SessionRowTree({ node, level }: { node: SessionNode; level: number }) {
  const { session, children, isRetry } = node;

  return (
    <>
      <tr className={`hover:bg-slate-800/40 transition-colors ${isRetry ? "bg-amber-950/20" : ""}`}>
        <td className="px-6 py-4 font-mono text-xs text-white whitespace-nowrap">
          <div className="flex items-center space-x-2" style={{ paddingLeft: `${level * 1.5}rem` }}>
            {isRetry && (
              <span className="text-amber-400 font-bold bg-amber-500/10 px-2 py-0.5 rounded text-[11px] border border-amber-500/30">
                ↳ Retry
              </span>
            )}
            <span className="font-semibold text-slate-100">{session.id}</span>
          </div>
        </td>
        <td className="px-6 py-4 font-mono text-xs text-slate-400 whitespace-nowrap">
          {session.topic_id}
        </td>
        <td className="px-6 py-4 whitespace-nowrap">
          <StatusBadge status={session.status} />
        </td>
        <td className="px-6 py-4 font-medium text-white whitespace-nowrap">{session.requested_count}</td>
        <td className="px-6 py-4 font-medium text-white whitespace-nowrap">{session.generated_count}</td>
        <td className="px-6 py-4 font-mono text-xs text-slate-300 whitespace-nowrap">{session.tokens_used.toLocaleString()}</td>
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

      {children.map((child) => (
        <SessionRowTree key={child.session.id} node={child} level={level + 1} />
      ))}
    </>
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
