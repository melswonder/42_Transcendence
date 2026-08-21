import { redirect } from "next/navigation";

import { AppShell } from "@/components/app-shell";
import { GameScreen } from "@/components/game-screen";
import { getCurrentUser } from "@/lib/auth";

/** 対局画面。盤面は WebSocket 越しにサーバーが配る全量 state だけで描く。
 * 中身が丸ごとクライアント側なのは、リアルタイム同期が
 * Server Component + 再フェッチの型に乗らないため。
 */
export default async function GamePage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return (
    <AppShell user={user}>
      <GameScreen />
    </AppShell>
  );
}
