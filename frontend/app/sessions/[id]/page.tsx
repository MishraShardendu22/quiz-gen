"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { api, Session } from "@/src/lib/api";
import { config } from "@/src/config";
import {
  saveRetryMapping,
  getParentSessionId,
  getChildSessionIds,
  getRetryMap,
} from "@/src/lib/retryTracker";

export default function SessionDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params?.id as string;

  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [parentSessionId, setParentSessionId] = useState<string | null>(null);
  const [childSessions, setChildSessions] = useState<Session[]>([]);

  const [showRetryModal, setShowRetryModal] = useState<boolean>(false);
  const [retryBudget, setRetryBudget] = useState<number>(10000);
  const [retrying, setRetrying] = useState<boolean>(false);

  useEffect(() => {
    if (!sessionId) return;

    let isSubscribed = true;
    let timerId: NodeJS.Timeout;

    async function pollSession() {
      try {
        const data = await api.getSession(sessionId);
        if (!isSubscribed) return;

        setSession(data);
        setLoading(false);

        const pId = getParentSessionId(sessionId);
        setParentSessionId(pId);

        try {
          const allSessions = await api.getSessions();
          if (allSessions && isSubscribed) {
            if (!pId) {
              const explicitParent = getRetryMap()[sessionId];
              if (explicitParent) {
                setParentSessionId(explicitParent);
              } else {
                const sameTopic = allSessions
                  .filter((s) => s.topic_id === data.topic_id)
                  .sort((a, b) => a.created_at - b.created_at);

                const currentIdx = sameTopic.findIndex((s) => s.id === sessionId);
                if (currentIdx > 0) {
                  for (let i = currentIdx - 1; i >= 0; i--) {
                    if (sameTopic[i].status === "failed") {
                      setParentSessionId(sameTopic[i].id);
                      saveRetryMapping(sameTopic[i].id, sessionId);
                      break;
                    }
                  }
                }
              }
            }

            const knownChildIds = new Set(getChildSessionIds(sessionId));
            const sameTopic = allSessions
              .filter((s) => s.topic_id === data.topic_id)
              .sort((a, b) => a.created_at - b.created_at);

            const currentIdx = sameTopic.findIndex((s) => s.id === sessionId);
            if (currentIdx >= 0 && data.status === "failed") {
              for (let i = currentIdx + 1; i < sameTopic.length; i++) {
                knownChildIds.add(sameTopic[i].id);
              }
            }

            const childrenList = allSessions
              .filter((s) => knownChildIds.has(s.id))
              .sort((a, b) => a.created_at - b.created_at);

            setChildSessions(childrenList);
          }
        } catch (e) {
          // Non-critical lookup exception
        }

        if (data.status === "pending" || data.status === "processing") {
          timerId = setTimeout(pollSession, config.pollingInterval);
        }
      } catch (err: any) {
        if (!isSubscribed) return;
        setError(err.message || "Failed to fetch session details");
        setLoading(false);
      }
    }

    pollSession();

    return () => {
      isSubscribed = false;
      if (timerId) clearTimeout(timerId);
    };
  }, [sessionId]);

  const handleRetrySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sessionId) return;
    setRetrying(true);

    try {
      const res = await api.retrySession(sessionId, retryBudget);
      saveRetryMapping(sessionId, res.session_id);
      setShowRetryModal(false);
      router.push(`/sessions/${res.session_id}`);
    } catch (err: any) {
      alert(err.message || "Failed to retry session");
      setRetrying(false);
    }
  };

  if (loading) {
    return (
      <div className="p-8 bg-slate-900 border border-slate-800 rounded-xl text-slate-400 text-sm">
        Loading session details...
      </div>
    );
  }

  if (error || !session) {
    return (
      <div className="space-y-4">
        <div className="p-4 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-sm">
          {error || "Session not found"}
        </div>
        <Link href="/sessions" className="text-indigo-400 hover:text-indigo-300 text-sm font-medium">
          ← Back to Sessions
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <Link href="/sessions" className="text-xs text-indigo-400 hover:text-indigo-300 mb-1 inline-block font-semibold">
            ← Back to Sessions
          </Link>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Session Overview</h1>
          <p className="font-mono text-xs text-slate-400 mt-1">ID: {session.id}</p>
        </div>

        <div className="flex items-center space-x-3">
          {session.status === "failed" && (
            <button
              onClick={() => {
                setRetryBudget(Math.max(session.token_budget * 2, session.tokens_used + 5000, 10000));
                setShowRetryModal(true);
              }}
              className="px-5 py-2.5 bg-gradient-to-r from-rose-600 to-red-600 hover:from-rose-500 hover:to-red-500 text-white rounded-xl text-xs font-bold shadow-lg shadow-rose-600/30 transition-all duration-200"
            >
              🔄 Retry Generation
            </button>
          )}
          <StatusBadge status={session.status} />
        </div>
      </div>

      {/* Metrics Cards Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</p>
          <div className="mt-2">
            <StatusBadge status={session.status} />
          </div>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Requested</p>
          <p className="text-2xl font-black text-white mt-1">{session.requested_count}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Generated</p>
          <p className="text-2xl font-black text-white mt-1">{session.generated_count}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Budget</p>
          <p className="text-xl font-black text-white mt-1 font-mono">{session.token_budget.toLocaleString()}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Tokens Used</p>
          <p className="text-xl font-black text-white mt-1 font-mono">{session.tokens_used.toLocaleString()}</p>
        </div>

        <div className="bg-slate-900/90 p-5 rounded-xl border border-slate-800 shadow-xl backdrop-blur-md">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Created At</p>
          <p className="text-xs text-slate-300 mt-2 font-semibold">
            {new Date(session.created_at * 1000).toLocaleTimeString()}
          </p>
        </div>
      </div>

      {/* Retry Relationship Card */}
      {(parentSessionId || childSessions.length > 0) && (
        <div className="bg-gradient-to-r from-amber-950/30 via-slate-900 to-slate-900 border border-amber-500/30 rounded-xl p-5 space-y-4 shadow-xl">
          {parentSessionId && (
            <div className="flex items-center space-x-3 text-sm text-amber-200">
              <span className="font-bold text-amber-400">Retry Of:</span>
              <Link
                href={`/sessions/${parentSessionId}`}
                className="font-mono text-xs text-indigo-300 hover:text-white font-bold bg-slate-800 px-3 py-1 rounded-lg border border-slate-700 transition-colors"
              >
                {parentSessionId}
              </Link>
            </div>
          )}

          {childSessions.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-bold text-amber-300">Retry Information</h3>
              <div className="space-y-2">
                {childSessions.map((cSession) => (
                  <div
                    key={cSession.id}
                    className="flex flex-col sm:flex-row items-start sm:items-center justify-between bg-slate-800/80 p-3.5 rounded-lg border border-slate-700 text-xs gap-3"
                  >
                    <div className="flex items-center space-x-3">
                      <span className="font-semibold text-slate-300">Retry Session:</span>
                      <span className="font-mono text-white font-bold">{cSession.id}</span>
                      <StatusBadge status={cSession.status} />
                    </div>
                    <Link
                      href={`/sessions/${cSession.id}`}
                      className="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-bold transition-colors"
                    >
                      Open Retry
                    </Link>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Processing Banner */}
      {(session.status === "pending" || session.status === "processing") && (
        <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl text-amber-300 text-sm flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <span className="animate-spin text-amber-400 font-bold text-lg">↻</span>
            <span className="font-medium">Quiz generation in progress... Polling status every 2 seconds.</span>
          </div>
        </div>
      )}

      {/* Error Details Card */}
      {session.error && (
        <div className="p-5 bg-rose-950/40 border border-rose-800/60 rounded-xl">
          <h3 className="text-xs font-bold text-rose-400 uppercase tracking-wider mb-1">Error Details</h3>
          <p className="text-sm text-rose-200 font-medium">{session.error}</p>
        </div>
      )}

      {/* Generated Questions Section */}
      <div className="space-y-4">
        <h2 className="text-xl font-bold text-white">
          Generated Questions ({session.questions ? session.questions.length : 0})
        </h2>

        {session.status === "completed" && (!session.questions || session.questions.length === 0) && (
          <div className="p-6 bg-slate-900 border border-slate-800 rounded-xl text-slate-400 text-sm">
            Session completed, but no questions were returned.
          </div>
        )}

        {session.questions && session.questions.length > 0 && (
          <div className="space-y-4">
            {session.questions.map((q, idx) => (
              <div key={q.id || idx} className="bg-slate-900/90 p-6 rounded-xl border border-slate-800 shadow-xl space-y-4 backdrop-blur-md">
                <div className="flex items-start justify-between">
                  <h3 className="text-base font-bold text-white">
                    {idx + 1}. {q.question}
                  </h3>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                  {[q.option_1, q.option_2, q.option_3, q.option_4].map((option, optIdx) => {
                    const isCorrect = q.correct_answer === optIdx;
                    return (
                      <div
                        key={optIdx}
                        className={`p-3.5 rounded-xl border text-sm flex items-center justify-between transition-all ${
                          isCorrect
                            ? "bg-emerald-950/40 border-emerald-500/50 text-emerald-200 font-semibold"
                            : "bg-slate-800/40 border-slate-800 text-slate-300"
                        }`}
                      >
                        <span>
                          <strong className="mr-2 text-slate-400">{String.fromCharCode(65 + optIdx)}.</strong>
                          {option}
                        </span>
                        {isCorrect && (
                          <span className="text-[11px] bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 px-2.5 py-0.5 rounded-full font-bold">
                            ✓ Correct
                          </span>
                        )}
                      </div>
                    );
                  })}
                </div>

                {q.explanation && (
                  <div className="mt-3 pt-3 border-t border-slate-800 text-xs text-slate-400">
                    <strong className="text-slate-300">Explanation: </strong>
                    {q.explanation}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Retry Modal */}
      {showRetryModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm">
          <div className="w-full max-w-md bg-slate-900 rounded-2xl shadow-2xl border border-slate-800 p-6 space-y-4">
            <h2 className="text-xl font-bold text-white">Retry Quiz Generation</h2>
            <p className="text-xs text-slate-400">
              Creates a new session for topic <code className="text-indigo-300">{session.topic_id}</code> with requested count of <code className="text-indigo-300">{session.requested_count}</code> questions.
            </p>

            <form onSubmit={handleRetrySubmit} className="space-y-4 pt-2">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                  New Token Budget
                </label>
                <input
                  type="number"
                  min={100}
                  value={retryBudget}
                  onChange={(e) => setRetryBudget(parseInt(e.target.value) || 100)}
                  required
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl px-4 py-2.5 text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-rose-500"
                />
                <p className="text-xs text-slate-500 mt-1">Previous budget: {session.token_budget} | Tokens used: {session.tokens_used}</p>
              </div>

              <div className="flex justify-end space-x-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setShowRetryModal(false)}
                  disabled={retrying}
                  className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl text-xs font-semibold transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={retrying}
                  className="px-4 py-2 bg-rose-600 hover:bg-rose-500 text-white rounded-xl text-xs font-bold shadow-lg shadow-rose-600/30 transition-all disabled:opacity-50"
                >
                  {retrying ? "Creating..." : "Start Retry Session"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
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
    <span className={`inline-flex px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider border ${colorClass}`}>
      {status}
    </span>
  );
}
