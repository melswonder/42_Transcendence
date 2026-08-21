import { redirect } from "next/navigation";
import { Group, Stack, Text, Title } from "@mantine/core";

import { AchievementList } from "@/components/achievement-list";
import { AppShell } from "@/components/app-shell";
import { StatsStream } from "@/components/use-stats-stream";
import { getCurrentUser } from "@/lib/auth";
import { getAchievements } from "@/lib/stats";

export default async function AchievementsPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  const achievements = await getAchievements();

  return (
    <AppShell user={user}>
      <StatsStream />
      <Stack gap="lg" maw={1000} mx="auto">
        <Group justify="space-between" align="baseline">
          <Title order={2}>実績</Title>
          <Text c="dimmed">
            {achievements.unlocked_count} / {achievements.total_count} 解除
          </Text>
        </Group>

        <AchievementList items={achievements.items} />
      </Stack>
    </AppShell>
  );
}
