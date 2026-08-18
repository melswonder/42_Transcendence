import { GoogleLoginButton } from "./google-login-button";

// バックエンドがログイン失敗時に付けてくるクエリ。
// 値は handler/auth.go の redirectWithError が渡す文字列と一対一で対応する。
const ERROR_MESSAGES: Record<string, string> = {
  access_denied: "Google の同意画面でキャンセルされました。",
  invalid_state:
    "ログインの照合に失敗しました。時間が経ちすぎた可能性があります。もう一度お試しください。",
  missing_code: "認可コードを受け取れませんでした。もう一度お試しください。",
  login_failed: "ログイン処理に失敗しました。しばらくしてからお試しください。",
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;
  // 未知の値が来ても素性を晒さず、一般的な文言に丸める。
  const message = error
    ? (ERROR_MESSAGES[error] ?? "ログインできませんでした。")
    : null;

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-900 p-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-800/80 p-8 shadow-2xl backdrop-blur-xl">
        <div className="mb-8 text-center">
          <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-xl bg-emerald-500/20 text-emerald-400">
            <PongMark />
          </div>
          <h1 className="text-3xl font-bold tracking-wider text-white">
            TRANSCENDENCE
          </h1>
          <p className="mt-2 text-slate-400">Sign in to play</p>
        </div>

        {message && (
          <div
            role="alert"
            className="mb-6 flex items-start gap-3 rounded-lg border border-rose-500/30 bg-rose-500/10 p-4 text-sm text-rose-300"
          >
            <AlertMark />
            <span>{message}</span>
          </div>
        )}

        <GoogleLoginButton />

        <p className="mt-6 text-center text-xs text-slate-500">
          続行すると、Google
          アカウントの表示名・メールアドレス・プロフィール画像を取得します。
        </p>
      </div>
    </div>
  );
}

// ロゴ代わりのパドルとボール。lucide-react を足さずに済ませるためのインライン SVG。
function PongMark() {
  return (
    <svg
      width="32"
      height="32"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M4 7v10" />
      <path d="M20 7v10" />
      <circle cx="12" cy="12" r="2" fill="currentColor" stroke="none" />
    </svg>
  );
}

function AlertMark() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      className="mt-0.5 shrink-0"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="9" />
      <path d="M12 8v4" />
      <path d="M12 16h.01" />
    </svg>
  );
}
