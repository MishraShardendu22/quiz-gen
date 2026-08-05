"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api, Session } from "@/src/lib/api";
import { config } from "@/src/config";

export default function SessionDetailsPage() {
  const params = useParams();
  const sessionId = params?.id as string;

  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

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

        // Keep polling if session is pending or processing
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

  if (loading) {
    return (
      <div className="p-8 bg-white border border-gray-200 rounded-lg shadow-sm">
        <p className="text-gray-500 text-sm">Loading session details...</p>
      </div>
    );
  }

  if (error || !session) {
    return (
      <div className="space-y-4">
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error || "Session not found"}
        </div>
        <Link href="/sessions" className="text-blue-600 hover:text-blue-800 text-sm font-medium">
          ← Back to Sessions
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <Link href="/sessions" className="text-xs text-blue-600 hover:text-blue-800 mb-1 inline-block">
            ← Back to Sessions
          </Link>
          <h1 className="text-2xl font-bold text-gray-900">Session Details</h1>
          <p className="font-mono text-xs text-gray-500 mt-1">ID: {session.id}</p>
        </div>

        <StatusBadge status={session.status} />
      </div>

      {/* Overview Card */}
      <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm grid grid-cols-1 sm:grid-cols-4 gap-4">
        <div>
          <p className="text-xs font-medium text-gray-500 uppercase">Status</p>
          <p className="text-lg font-bold text-gray-900 capitalize mt-1">{session.status}</p>
        </div>

        <div>
          <p className="text-xs font-medium text-gray-500 uppercase">Requested Count</p>
          <p className="text-lg font-bold text-gray-900 mt-1">{session.requested_count}</p>
        </div>

        <div>
          <p className="text-xs font-medium text-gray-500 uppercase">Generated Count</p>
          <p className="text-lg font-bold text-gray-900 mt-1">{session.generated_count}</p>
        </div>

        <div>
          <p className="text-xs font-medium text-gray-500 uppercase">Topic ID</p>
          <p className="font-mono text-xs text-gray-900 truncate mt-1">{session.topic_id}</p>
        </div>
      </div>

      {/* Processing Banner */}
      {(session.status === "pending" || session.status === "processing") && (
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-md text-blue-800 text-sm flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <span className="animate-spin text-blue-600 font-bold">↻</span>
            <span>Quiz generation in progress... Polling status every 2 seconds.</span>
          </div>
        </div>
      )}

      {/* Error Message Display */}
      {session.error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <h3 className="text-sm font-semibold text-red-800">Generation Error</h3>
          <p className="text-sm text-red-700 mt-1">{session.error}</p>
        </div>
      )}

      {/* Questions Section */}
      <div className="space-y-4">
        <h2 className="text-xl font-bold text-gray-900">
          Generated Questions ({session.questions ? session.questions.length : 0})
        </h2>

        {session.status === "completed" && (!session.questions || session.questions.length === 0) && (
          <div className="p-6 bg-white border border-gray-200 rounded-lg text-gray-500 text-sm">
            Session completed, but no questions were returned.
          </div>
        )}

        {session.questions && session.questions.length > 0 && (
          <div className="space-y-4">
            {session.questions.map((q, idx) => (
              <div key={q.id || idx} className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm space-y-3">
                <div className="flex items-start justify-between">
                  <h3 className="text-base font-semibold text-gray-900">
                    {idx + 1}. {q.question}
                  </h3>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-2">
                  {[q.option_1, q.option_2, q.option_3, q.option_4].map((option, optIdx) => {
                    const isCorrect = q.correct_answer === optIdx;
                    return (
                      <div
                        key={optIdx}
                        className={`p-3 rounded border text-sm flex items-center justify-between ${
                          isCorrect
                            ? "bg-green-50 border-green-300 text-green-900 font-medium"
                            : "bg-gray-50 border-gray-200 text-gray-700"
                        }`}
                      >
                        <span>
                          <strong className="mr-2">{String.fromCharCode(65 + optIdx)}.</strong>
                          {option}
                        </span>
                        {isCorrect && (
                          <span className="text-xs bg-green-200 text-green-800 px-2 py-0.5 rounded font-semibold">
                            Correct
                          </span>
                        )}
                      </div>
                    );
                  })}
                </div>

                {q.explanation && (
                  <div className="mt-3 pt-3 border-t border-gray-100 text-xs text-gray-600">
                    <strong className="text-gray-800">Explanation: </strong>
                    {q.explanation}
                  </div>
                )}
              </div>
            ))}
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
    <span className={`inline-flex px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider ${colorClass}`}>
      {status}
    </span>
  );
}
