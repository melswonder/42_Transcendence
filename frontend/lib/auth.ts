import { cookies } from "next/headers";

import { serverApiUrl } from "@/lib/api";

/** GET /auth/me のレスポンス。backend の handler/auth.go userResponse と対。 */
export interface User {
  id: string;
  email: string | null;
  display_name: string;
  handle: string;
  level: number;
}

/** ログイン中のユーザーを返す。未ログインなら null。
 *
 * Server Component からの fetch には Cookie が自動では乗らないので、
 * ブラウザから届いたものをそのまま backend へ転送する。
 * セッション Cookie は :4000 が発行したものだが、Cookie はポートを区別しないので
 * :3000 にも届いている。
 */
export async function getCurrentUser(): Promise<User | null> {
  const cookie = (await cookies()).toString();
  if (!cookie) return null;

  const res = await fetch(serverApiUrl("/auth/me"), {
    headers: { cookie },
    cache: "no-store",
  });
  if (!res.ok) return null;

  return res.json();
}
