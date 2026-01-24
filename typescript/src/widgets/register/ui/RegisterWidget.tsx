import Link from 'next/link';

export function RegisterWidget() {
  return (
    <div>
      <h1>ユーザー登録</h1>
      <p>登録フォームがここに表示されます</p>
      <nav>
        <ul>
          <li>
            <Link href="/home">登録後ゲームホームへ</Link>
          </li>
          <li>
            <Link href="/login">ログインはこちら</Link>
          </li>
          <li>
            <Link href="/">トップに戻る</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
