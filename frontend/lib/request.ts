import { apiUrl } from "@/lib/api";

/** backend の失敗レスポンス。code は翻訳の鍵（無い場合もある）。 */
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string | undefined,
    message: string,
  ) {
    super(message);
  }
}

/** Cookie 付きで backend を叩き、失敗は ApiError に揃える。
 * 表示するときは code を messages の apiErrors から引き、無ければ message を出す。
 */
export async function requestJSON(path: string, init?: RequestInit) {
  const res = await fetch(apiUrl(path), { credentials: "include", ...init });
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as {
      error?: string;
      code?: string;
    } | null;
    // code の無い 401 はセッション切れ。画面ごとに生の "unauthorized" を
    // 出しても打つ手が無いので、ログインへ送り返す。
    if (res.status === 401 && !body?.code && typeof window !== "undefined") {
      window.location.href = "/login";
    }
    throw new ApiError(
      res.status,
      body?.code ?? (res.status === 401 ? "unauthorized" : undefined),
      body?.error ?? `HTTP ${res.status}`,
    );
  }
  return res.status === 204 ? null : res.json();
}
