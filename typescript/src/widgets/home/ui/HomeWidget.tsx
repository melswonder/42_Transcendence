'use client';

import Link from 'next/link';
import API from '@/app/api/page';

export function HomeWidget() {
  return (
    <div 
      className="relative min-h-screen flex items-center justify-center"
      style={{
        backgroundImage: 'url(/background.png)',
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
      }}
    >

      {/* メインコンテンツ */}
      <div className="w-full max-w-md mx-4">
        <div className="bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl p-8">
          <h1 className="text-4xl font-bold text-center mb-8 text-gray-800">
            42 Transcendence
          </h1>

          {/* 主要ボタン */}
          <div className="space-y-3 mb-6">
            <Link
              href="/login"
              className="block w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-4 rounded-lg text-center transition-colors"
            >
              ログイン
            </Link>
            <Link
              href="/register"
              className="block w-full bg-white hover:bg-gray-50 text-blue-600 font-semibold py-3 px-4 rounded-lg text-center border-2 border-blue-600 transition-colors"
            >
              新規登録
            </Link>
          </div>

          {/* ゲームメニュー */}
          <div className="space-y-2 mb-6">
            <Link
              href="/home"
              className="block w-full bg-blue-100 hover:bg-blue-200 text-blue-800 font-medium py-2.5 px-4 rounded-lg text-center transition-colors"
            >
              ゲームホーム
            </Link>
            <Link
              href="/home/matching"
              className="block w-full bg-blue-100 hover:bg-blue-200 text-blue-800 font-medium py-2.5 px-4 rounded-lg text-center transition-colors"
            >
              マッチング
            </Link>
            <Link
              href="/home/room"
              className="block w-full bg-blue-100 hover:bg-blue-200 text-blue-800 font-medium py-2.5 px-4 rounded-lg text-center transition-colors"
            >
              ルーム一覧
            </Link>
          </div>

          {/* フッターリンク */}
          <div className="space-y-1 pt-4 border-t border-gray-200">
            <Link
              href="/legal/privacy"
              className="block w-full text-gray-600 hover:text-gray-800 text-sm py-2 px-4 text-center transition-colors"
            >
              プライバシーポリシー
            </Link>
            <Link
              href="/legal/user-agreement"
              className="block w-full text-gray-600 hover:text-gray-800 text-sm py-2 px-4 text-center transition-colors"
            >
              利用規約
            </Link>
            <API />
          </div>
        </div>
      </div>
    </div>
  );
}
