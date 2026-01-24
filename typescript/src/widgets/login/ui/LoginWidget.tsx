import Link from 'next/link';

export function LoginWidget() {
  return (
    <div>
      <h1>ログイン</h1>
      <p>ログインフォームがここに表示されます</p>
      <nav>
        <ul>
          <li>
            <Link href="/home">ログイン後ゲームホームへ</Link>
          </li>
          <li>
            <Link href="/register">新規登録はこちら</Link>
          </li>
          <li>
            <Link href="/">トップに戻る</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
