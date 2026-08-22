import { redirect } from "next/navigation";
import { Stack, Title } from "@mantine/core";

import { AppShell } from "@/components/app-shell";
import { FriendsScreen } from "@/components/friends-screen";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";

/** フレンド画面。一覧・申請の送受・ユーザー検索。
 * 申請や承認で状態がすぐ変わるので、Server Component ではなく
 * クライアント側でフェッチして操作のたびに引き直す。
 */
export default async function FriendsPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");
  const t = await getTranslations("friends");

  return (
    <AppShell user={user}>
      <Stack gap="lg" maw={720} mx="auto">
        <Title order={2} size="h3">
          {t("title")}
        </Title>
        <FriendsScreen />
      </Stack>
    </AppShell>
  );
}
