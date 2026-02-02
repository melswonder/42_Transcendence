'use client';

import Link from 'next/link';

export function HomeWidget() {
  return (
    <div 
      className="relative min-h-screen flex items-center justify-center p-8"
      style={{
        backgroundImage: 'url(/background.png)',
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
      }}
    >
      {/* メインコンテンツ */}
      <div className="flex flex-col md:flex-row items-center gap-12 max-w-6xl w-full">
        {/* 左側: GIF画像を配置するコンテナ */}
        <div className="w-full max-w-[500px] aspect-square relative">
          <img 
            src="/chess-board-animation.gif" 
            alt="Chess Animation"
            className="rounded-md w-full h-full object-contain"
          />
        </div>

        {/* 右側: テキストとアクションボタン */}
        <div className="flex flex-col items-center md:items-start text-center md:text-left max-w-[400px]">
          <h1 className="text-4xl md:text-5xl font-extrabold leading-tight mb-4 text-white drop-shadow-lg">
            ルールはない戦え
          </h1>
          
          <p className="text-gray-200 text-lg mb-8 drop-shadow">
            将棋とチェス、二つの伝統が交わる場所で、新たな伝説の一手を。
          </p>

          <Link 
            href="/login"
            className="bg-[#81b64c] hover:bg-[#95c95d] text-white font-bold py-4 px-12 rounded-lg text-2xl shadow-lg transition-transform active:scale-95"
          >
            Get Started
          </Link>
        </div>
      </div>
    </div>
  );
}
