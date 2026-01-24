import Link from 'next/link';

interface RoomDetailWidgetProps {
  roomId: string;
}

export function RoomDetailWidget({ roomId }: RoomDetailWidgetProps) {
  return (
    <div>
      <h1>ルーム: {roomId}</h1>
      <p>対戦相手を待っています...</p>
      <nav>
        <ul>
          <li>
            <Link href={`/battle/${roomId}-battle`}>バトル開始</Link>
          </li>
          <li>
            <Link href="/home/room">ルーム一覧に戻る</Link>
          </li>
          <li>
            <Link href="/home">ホームに戻る</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
