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
        <h1 className="text-2xl font-bold text-gray-900">Topics</h1>
        <p className="text-gray-600 text-sm">Select a topic to generate quiz questions</p>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-6 text-gray-500 text-sm">Loading topics...</div>
        ) : topics.length === 0 ? (
          <div className="p-6 text-gray-500 text-sm">No topics available.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-gray-600">
              <thead className="bg-gray-50 text-gray-700 font-medium border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3">Topic Name</th>
                  <th className="px-6 py-3">Document Count</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {topics.map((topic) => (
                  <tr key={topic.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-semibold text-gray-900">
                      {topic.name}
                    </td>
                    <td className="px-6 py-4">{topic.document_count}</td>
                    <td className="px-6 py-4">
                      <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 capitalize">
                        {topic.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button
                        onClick={() => setSelectedTopic(topic)}
                        className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 text-xs font-medium"
                      >
                        Generate Questions
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
