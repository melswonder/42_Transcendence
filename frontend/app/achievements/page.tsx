import { redirect } from "next/navigation";
import { Group, Stack, Text, Title } from "@mantine/core";

import { AchievementList } from "@/components/achievement-list";
import { AppShell } from "@/components/app-shell";
import { StatsStream } from "@/components/use-stats-stream";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";
import { getAchievements } from "@/lib/stats";

export default async function AchievementsPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  const achievements = await getAchievements();
  const t = await getTranslations("achievements");

  return (
    <AppShell user={user}>
      <StatsStream />
      <Stack gap="lg" maw={1000} mx="auto">
        <Group justify="space-between" align="baseline">
          <Title order={2}>{t("title")}</Title>
          <Text c="dimmed">
            {t("unlockedCount", {
              unlocked: achievements.unlocked_count,
              total: achievements.total_count,
            })}
          </Text>
        </Group>

        <AchievementList items={achievements.items} />
      </Stack>
    </AppShell>
  );
}
