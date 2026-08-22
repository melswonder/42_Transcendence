import { redirect } from "next/navigation";

import { AppShell } from "@/components/app-shell";
import { WatchScreen } from "@/components/watch-screen";
import { getCurrentUser } from "@/lib/auth";

/** 観戦画面。進行中の対局一覧と、選んだ対局のライブ観戦。
 * 盤面は対局者と同じ WebSocket の全量 state で描くので、
 * 遅延や再接続への強さも対局画面と同じになる。
 */
export default async function WatchPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return (
    <AppShell user={user}>
      <WatchScreen />
    </AppShell>
  );
}
