const stripTrailingSlash = (base: string) => base.replace(/\/+$/, "");

/** ブラウザから見たバックエンドの入口。フロント自身の :3000 ではないので注意。 */
export const API_BASE = stripTrailingSlash(
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000",
);

/** Server Component / Route Handler から見た入口。
 * コンテナの中では localhost:4000 が自分自身を指してしまうので、
 * ブラウザ用とは別に持つ（compose が API_URL=http://backend:4000 を渡す）。
 */
const SERVER_API_BASE = stripTrailingSlash(
  process.env.API_URL ?? "http://localhost:4000",
);

/** パスを絶対 URL にする。先頭スラッシュの有無は問わない。 */
export function apiUrl(path: string): string {
  return `${API_BASE}/${path.replace(/^\/+/, "")}`;
}

/** サーバー側から呼ぶ版の apiUrl。 */
export function serverApiUrl(path: string): string {
  return `${SERVER_API_BASE}/${path.replace(/^\/+/, "")}`;
}

/** WebSocket 用。http(s) を ws(s) に読み替える。ブラウザからしか使わない。 */
export function wsUrl(path: string): string {
  return apiUrl(path).replace(/^http/, "ws");
}
