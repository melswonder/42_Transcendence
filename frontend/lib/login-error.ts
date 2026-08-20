/** キーは backend の handler/auth.go redirectWithError が渡す値と一対一。 */
const ERROR_MESSAGES: Record<string, string> = {
  access_denied: "Google の同意画面でキャンセルされました。",
  invalid_state:
    "ログインの照合に失敗しました。時間が経ちすぎた可能性があります。もう一度お試しください。",
  missing_code: "認可コードを受け取れませんでした。もう一度お試しください。",
  login_failed: "ログイン処理に失敗しました。しばらくしてからお試しください。",
};

/** 未知の値が来ても素性を晒さず、一般的な文言に丸める。 */
export function resolveLoginError(code: string | undefined): string | null {
  if (!code) return null;
  return ERROR_MESSAGES[code] ?? "ログインできませんでした。";
}
