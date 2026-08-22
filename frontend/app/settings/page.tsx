import { redirect } from "next/navigation";
import { Stack, Title } from "@mantine/core";

import { AppShell } from "@/components/app-shell";
import { SettingsForm } from "@/components/settings-form";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";

/** 設定画面。プロフィールの編集とアバターの管理。 */
export default async function SettingsPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");
  const t = await getTranslations("settings");

  return (
    <AppShell user={user}>
      <Stack gap="lg" maw={900} mx="auto">
        <Title order={2} size="h3">
          {t("title")}
        </Title>
        <SettingsForm user={user} />
      </Stack>
    </AppShell>
  );
}
