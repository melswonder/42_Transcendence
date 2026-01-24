'use client';

import Link from 'next/link';

export default function HomePage() {
  return (
    <div>
      <h1>ゲームホーム</h1>
      <nav>
        <ul>
          <li>
            <Link href="/home/matching">マッチング</Link>
          </li>
          <li>
            <Link href="/home/room">ルーム一覧</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
