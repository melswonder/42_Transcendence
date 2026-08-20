/** バックエンド（:4000）の入口。フロント自身の :3000 ではないので注意。
 * 末尾スラッシュ付きで設定されても "//auth/google" にならないよう落とす。
 */
export const API_BASE = (
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000"
).replace(/\/+$/, "");

/** パスを絶対 URL にする。先頭スラッシュの有無は問わない。 */
export function apiUrl(path: string): string {
  return `${API_BASE}/${path.replace(/^\/+/, "")}`;
}
