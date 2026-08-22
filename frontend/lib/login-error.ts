/** キーは backend の handler/auth.go redirectWithError が渡す値と一対一。 */
const KNOWN_ERRORS = new Set([
  "access_denied",
  "invalid_state",
  "missing_code",
  "login_failed",
]);

/** 翻訳キーに解決する。未知の値は素性を晒さず fallback に丸める。
 * 文言そのものは messages/​*.json の login.errors 配下にある。
 */
export function resolveLoginErrorKey(code: string | undefined): string | null {
  if (!code) return null;
  return KNOWN_ERRORS.has(code) ? code : "fallback";
}
