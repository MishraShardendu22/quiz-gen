"use client";

import { useEffect, useState } from "react";
import { api, Topic } from "@/src/lib/api";
import { GenerateModal } from "../components/GenerateModal";

export default function TopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [selectedTopic, setSelectedTopic] = useState<Topic | null>(null);

  useEffect(() => {
    async function loadTopics() {
      try {
        const data = await api.getTopics();
        setTopics(data || []);
      } catch (err: any) {
        setError(err.message || "Failed to load topics");
      } finally {
        setLoading(false);
      }
    }
    loadTopics();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Discovered Topics</h1>
        <p className="text-slate-400 text-sm mt-1">Select a markdown content topic to generate quiz questions</p>
      </div>

      {error && (
        <div className="p-4 bg-rose-950/50 border border-rose-800/80 rounded-xl text-rose-300 text-sm">
          {error}
        </div>
      )}

      <div className="bg-slate-900/90 rounded-xl border border-slate-800 shadow-xl overflow-hidden backdrop-blur-md">
        {loading ? (
          <div className="p-6 text-slate-400 text-sm">Loading topics...</div>
        ) : topics.length === 0 ? (
          <div className="p-6 text-slate-400 text-sm">No topics available.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300 min-w-[600px]">
              <thead className="bg-slate-800/60 text-slate-400 font-semibold tracking-wider uppercase text-xs border-b border-slate-800">
                <tr>
                  <th className="px-6 py-3.5">Topic Name</th>
                  <th className="px-6 py-3.5">Document Count</th>
                  <th className="px-6 py-3.5">Status</th>
                  <th className="px-6 py-3.5 text-right whitespace-nowrap">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {topics.map((topic) => (
                  <tr key={topic.id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="px-6 py-4 font-bold text-white">
                      {topic.name}
                    </td>
                    <td className="px-6 py-4 font-medium text-slate-300">{topic.document_count}</td>
                    <td className="px-6 py-4">
                      <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 capitalize">
                        {topic.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right whitespace-nowrap">
                      <button
                        onClick={() => setSelectedTopic(topic)}
                        className="px-4 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-lg text-xs font-bold shadow-lg shadow-indigo-600/20 transition-all whitespace-nowrap"
                      >
                        ⚡ Generate Questions
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {selectedTopic && (
        <GenerateModal
          topic={selectedTopic}
          onClose={() => setSelectedTopic(null)}
        />
      )}
    </div>
  );
}
