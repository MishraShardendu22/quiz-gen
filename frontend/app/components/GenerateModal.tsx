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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-md bg-white rounded-lg shadow-lg border border-gray-200 p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold text-gray-900">Generate Quiz Questions</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 font-semibold text-lg"
          >
            ✕
          </button>
        </div>

        <p className="text-sm text-gray-600 mb-6">
          Topic: <span className="font-semibold text-gray-900">{topic.name}</span>
        </p>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Question Count (1-100)
            </label>
            <input
              type="number"
              min={1}
              max={100}
              value={requestedCount}
              onChange={(e) => setRequestedCount(parseInt(e.target.value) || 1)}
              required
              className="w-full border border-gray-300 rounded px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Token Budget
            </label>
            <input
              type="number"
              min={100}
              value={tokenBudget}
              onChange={(e) => setTokenBudget(parseInt(e.target.value) || 100)}
              required
              className="w-full border border-gray-300 rounded px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-4 border-t border-gray-100">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 border border-gray-300 rounded text-gray-700 hover:bg-gray-50 text-sm font-medium disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm font-medium disabled:opacity-50"
            >
              {loading ? "Starting..." : "Generate"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
