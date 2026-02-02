"use client";

import React, { useState } from "react";

interface ApiTest {
  message: string;
}

export default function API() {
  const [data, setData] = useState<ApiTest>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000");
      if (!response.ok) {
        throw new Error("ネットワーク応答が正常ではありませんでした");
      }
      const result = await response.json();
      setData(result);
    } catch (error) {
      console.error("エラー発生:", error);
      setError(error instanceof Error ? error.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">API データ</h1>
      <button
        onClick={fetchData}
        disabled={loading}
        className={`px-6 py-3 text-lg font-semibold rounded-lg mb-6 transition-colors ${
          loading
            ? "bg-gray-400 cursor-not-allowed"
            : "bg-blue-600 hover:bg-blue-700 text-white cursor-pointer"
        }`}
      >
        {loading ? "読み込み中..." : "APIを呼び出す"}
      </button>

      {error && (
        <div className="text-red-600 bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
          エラー: {error}
        </div>
      )}
      {data && (
        <pre className="bg-gray-100 p-4 rounded-lg overflow-auto text-sm">
          {JSON.stringify(data, null, 2)}
        </pre>
      )}
    </div>
  );
}
