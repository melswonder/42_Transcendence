import Link from 'next/link';

export function RoomWidget() {
  return (
    <div>
      <h1>ルーム一覧</h1>
      <nav>
        <ul>
          <li>
            <Link href="/home/room/room-001">ルーム 001 に参加</Link>
          </li>
          <li>
            <Link href="/home/room/room-002">ルーム 002 に参加</Link>
          </li>
          <li>
            <Link href="/home">ホームに戻る</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
