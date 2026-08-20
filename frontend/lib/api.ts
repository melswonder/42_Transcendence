/** バックエンドの入口。docker-compose が NEXT_PUBLIC_API_URL を渡している。
 *
 * フロント自身（:3000）ではなくバックエンド（:4000）を指すこと。
 * 末尾スラッシュ付きで設定されても "//auth/google" にならないよう落としておく。
 */
export const API_BASE = (
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000"
).replace(/\/+$/, "");

/** API のパスを絶対 URL に変換する。先頭スラッシュの有無は問わない。 */
export function apiUrl(path: string): string {
  return `${API_BASE}/${path.replace(/^\/+/, "")}`;
}
