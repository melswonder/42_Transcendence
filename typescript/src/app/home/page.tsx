'use client';

import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-8">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-2xl shadow-xl p-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">ゲームホーム</h1>
          <p className="text-gray-600 mb-8">ようこそ！プレイモードを選択してください</p>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* マッチング */}
            <Link
              href="/home/matching"
              className="group block p-6 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl shadow-lg hover:shadow-2xl transition-all hover:scale-105"
            >
              <div className="text-white">
                <h2 className="text-2xl font-bold mb-2">マッチング</h2>
                <p className="text-blue-100">ランダムな対戦相手と対戦</p>
              </div>
            </Link>

            {/* ルーム一覧 */}
            <Link
              href="/home/room"
              className="group block p-6 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg hover:shadow-2xl transition-all hover:scale-105"
            >
              <div className="text-white">
                <h2 className="text-2xl font-bold mb-2">ルーム一覧</h2>
                <p className="text-purple-100">既存のルームを探して参加</p>
              </div>
            </Link>
          </div>

          {/* ナビゲーション */}
          <div className="mt-8 pt-6 border-t border-gray-200">
            <Link
              href="/"
              className="text-blue-600 hover:text-blue-800 font-medium"
            >
              ← トップに戻る
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
