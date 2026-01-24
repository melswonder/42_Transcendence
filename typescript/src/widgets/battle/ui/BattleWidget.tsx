import Link from 'next/link';

interface BattleWidgetProps {
  battleId: string;
}

export function BattleWidget({ battleId }: BattleWidgetProps) {
  return (
    <div>
      <h1>対戦中 - Battle ID: {battleId}</h1>
      <p>ゲーム画面がここに表示されます</p>
      <nav>
        <Link href="/home">ホームに戻る</Link>
      </nav>
    </div>
  );
}
