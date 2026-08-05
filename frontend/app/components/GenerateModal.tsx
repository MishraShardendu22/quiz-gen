"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { api, Topic } from "@/src/lib/api";

interface GenerateModalProps {
  topic: Topic;
  onClose: () => void;
}

export function GenerateModal({ topic, onClose }: GenerateModalProps) {
  const router = useRouter();
  const [requestedCount, setRequestedCount] = useState<number>(5);
  const [tokenBudget, setTokenBudget] = useState<number>(4000);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await api.createGenerate({
        topic_id: topic.id,
        requested_count: requestedCount,
        token_budget: tokenBudget,
      });

      onClose();
      router.push(`/sessions/${res.session_id}`);
    } catch (err: any) {
      setError(err.message || "Failed to start generation");
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-sm">
      <div className="w-full max-w-md bg-slate-900 rounded-2xl shadow-2xl border border-slate-800 p-6 space-y-4">
        <div className="flex justify-between items-center pb-2 border-b border-slate-800">
          <h2 className="text-xl font-bold text-white">Generate Quiz Questions</h2>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white font-bold text-lg"
          >
            ✕
          </button>
        </div>

        <p className="text-xs text-slate-400">
          Topic: <span className="font-bold text-indigo-300">{topic.name}</span>
        </p>

        {error && (
          <div className="p-3 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-xs">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
              Question Count (1-100)
            </label>
            <input
              type="number"
              min={1}
              max={100}
              value={requestedCount}
              onChange={(e) => setRequestedCount(parseInt(e.target.value) || 1)}
              required
              className="w-full bg-slate-800 border border-slate-700 rounded-xl px-4 py-2.5 text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
              Token Budget
            </label>
            <input
              type="number"
              min={100}
              value={tokenBudget}
              onChange={(e) => setTokenBudget(parseInt(e.target.value) || 100)}
              required
              className="w-full bg-slate-800 border border-slate-700 rounded-xl px-4 py-2.5 text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-4 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl text-xs font-semibold transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl text-xs font-bold shadow-lg shadow-indigo-600/30 transition-all disabled:opacity-50"
            >
              {loading ? "Starting..." : "Start Generation"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
