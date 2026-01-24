import Link from 'next/link';

export function HomeWidget() {
  return (
    <div>
      <h1>ホーム</h1>
      <nav>
        <ul>
          <li>
            <Link href="/home">ゲームホーム</Link>
          </li>
          <li>
            <Link href="/login">ログイン</Link>
          </li>
          <li>
            <Link href="/register">新規登録</Link>
          </li>
          <li>
            <Link href="/legal/privacy">プライバシーポリシー</Link>
          </li>
          <li>
            <Link href="/legal/user-agreement">利用規約</Link>
          </li>
        </ul>
      </nav>
    </div>
  );
}
