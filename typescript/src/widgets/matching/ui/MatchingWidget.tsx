import Link from 'next/link';

export function MatchingWidget() {
  return (
    <div>
      <h1>マッチング</h1>
      <p>対戦相手を探しています...</p>
      <nav>
        <ul>
          <li>
            <Link href="/battle/demo-battle-123">バトル開始（デモ）</Link>
          </li>
          <li>
            <Link href="/home">ホームに戻る</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
