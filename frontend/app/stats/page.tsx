import { redirect } from "next/navigation";
import { Group, Progress, SimpleGrid, Stack, Text, Title } from "@mantine/core";
import {
  IconChartLine,
  IconFlame,
  IconMedal,
  IconStars,
  IconSwords,
  IconTrophy,
} from "@tabler/icons-react";

import { ActivityChart } from "@/components/activity-chart";
import { AppShell } from "@/components/app-shell";
import { ExportButtons } from "@/components/export-buttons";
import { OutcomeChart } from "@/components/outcome-chart";
import { RatingChart } from "@/components/rating-chart";
import { StatsFilters } from "@/components/stats-filters";
import { StatTile } from "@/components/stat-tile";
import { getCurrentUser } from "@/lib/auth";
import {
  getBreakdown,
  getSummary,
  getTimeseries,
  type StatsFilters as Filters,
} from "@/lib/stats";

export default async function StatsPage({
  searchParams,
}: {
  searchParams: Promise<Filters>;
}) {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  const filters = await searchParams;

  // 3 本とも独立しているので直列に待たない。
  const [summary, timeseries, breakdown] = await Promise.all([
    getSummary(filters),
    getTimeseries(filters),
    getBreakdown(filters),
  ]);

  const xpInLevel = summary.xp - summary.xp_for_level;
  const xpNeeded = summary.xp_for_next_level - summary.xp_for_level;
  const streak = summary.current_streak;

  return (
    <AppShell user={user}>
      <Stack gap="lg" maw={1100} mx="auto">
        <Group justify="space-between" align="center">
          <Title order={2}>統計</Title>
          <ExportButtons />
        </Group>

        <div className="print:hidden">
          <StatsFilters withInterval />
        </div>

        <SimpleGrid cols={{ base: 1, xs: 2, md: 3, lg: 6 }} spacing="md">
          <StatTile
            icon={<IconSwords size={20} />}
            label="対戦数"
            value={`${summary.total_matches}`}
            hint={`${summary.wins}勝 ${summary.losses}敗 ${summary.draws}分`}
          />
          <StatTile
            icon={<IconTrophy size={20} />}
            label="勝率"
            value={`${(summary.win_rate * 100).toFixed(1)}%`}
            hint="引き分けも母数に含む"
          />
          <StatTile
            icon={<IconChartLine size={20} />}
            label="レーティング"
            value={`${summary.rating}`}
            hint="ランク戦のみ変動"
          />
          <StatTile
            icon={<IconMedal size={20} />}
            label="順位"
            value={`${summary.ranking}位`}
            hint={`${summary.total_players} 人中`}
          />
          <StatTile
            icon={<IconFlame size={20} />}
            label={streak >= 0 ? "連勝" : "連敗"}
            value={`${Math.abs(streak)}`}
            hint={`最長 ${summary.best_streak} 連勝`}
            tone={streak > 0 ? "positive" : streak < 0 ? "negative" : undefined}
          />
          <StatTile
            icon={<IconStars size={20} />}
            label="レベル"
            value={`Lv.${summary.level}`}
            hint={`${xpInLevel} / ${xpNeeded} XP`}
          />
        </SimpleGrid>

        <Stack gap={4}>
          <Group justify="space-between">
            <Text size="sm" c="dimmed">
              次のレベルまで
            </Text>
            <Text size="sm" c="dimmed" ff="monospace">
              {xpNeeded - xpInLevel} XP
            </Text>
          </Group>
          <Progress
            value={xpNeeded > 0 ? (xpInLevel / xpNeeded) * 100 : 0}
            size="md"
            radius="xl"
            aria-label="次のレベルまでの進捗"
          />
        </Stack>

        <RatingChart points={timeseries.points} />
        <ActivityChart points={timeseries.points} />

        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
          <OutcomeChart
            title="決着のつき方"
            slices={breakdown.by_result_type}
          />
          <OutcomeChart title="モード別" slices={breakdown.by_mode} />
        </SimpleGrid>
      </Stack>
    </AppShell>
  );
}
