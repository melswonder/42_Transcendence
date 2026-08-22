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
import { getTranslations } from "next-intl/server";

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

  const t = await getTranslations("stats");
  const xpInLevel = summary.xp - summary.xp_for_level;
  const xpNeeded = summary.xp_for_next_level - summary.xp_for_level;
  const streak = summary.current_streak;

  return (
    <AppShell user={user}>
      <Stack gap="lg" maw={1100} mx="auto">
        <Group justify="space-between" align="center">
          <Title order={2}>{t("title")}</Title>
          <ExportButtons />
        </Group>

        <div className="print:hidden">
          <StatsFilters withInterval />
        </div>

        <SimpleGrid cols={{ base: 1, xs: 2, md: 3, lg: 6 }} spacing="md">
          <StatTile
            icon={<IconSwords size={20} />}
            label={t("tiles.matches")}
            value={`${summary.total_matches}`}
            hint={t("tiles.matchesHint", {
              wins: summary.wins,
              losses: summary.losses,
              draws: summary.draws,
            })}
          />
          <StatTile
            icon={<IconTrophy size={20} />}
            label={t("tiles.winRate")}
            value={`${(summary.win_rate * 100).toFixed(1)}%`}
            hint={t("tiles.winRateHint")}
          />
          <StatTile
            icon={<IconChartLine size={20} />}
            label={t("tiles.rating")}
            value={`${summary.rating}`}
            hint={t("tiles.ratingHint")}
          />
          <StatTile
            icon={<IconMedal size={20} />}
            label={t("tiles.ranking")}
            value={t("tiles.rankingValue", { rank: summary.ranking })}
            hint={t("tiles.rankingHint", { total: summary.total_players })}
          />
          <StatTile
            icon={<IconFlame size={20} />}
            label={streak >= 0 ? t("tiles.winStreak") : t("tiles.loseStreak")}
            value={`${Math.abs(streak)}`}
            hint={t("tiles.streakHint", { count: summary.best_streak })}
            tone={streak > 0 ? "positive" : streak < 0 ? "negative" : undefined}
          />
          <StatTile
            icon={<IconStars size={20} />}
            label={t("tiles.level")}
            value={t("tiles.levelValue", { level: summary.level })}
            hint={t("tiles.levelHint", {
              current: xpInLevel,
              needed: xpNeeded,
            })}
          />
        </SimpleGrid>

        <Stack gap={4}>
          <Group justify="space-between">
            <Text size="sm" c="dimmed">
              {t("nextLevel")}
            </Text>
            <Text size="sm" c="dimmed" ff="monospace">
              {xpNeeded - xpInLevel} XP
            </Text>
          </Group>
          <Progress
            value={xpNeeded > 0 ? (xpInLevel / xpNeeded) * 100 : 0}
            size="md"
            radius="xl"
            aria-label={t("nextLevelAria")}
          />
        </Stack>

        <RatingChart points={timeseries.points} />
        <ActivityChart points={timeseries.points} />

        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
          <OutcomeChart
            title={t("byResultType")}
            slices={breakdown.by_result_type}
          />
          <OutcomeChart title={t("byMode")} slices={breakdown.by_mode} />
        </SimpleGrid>
      </Stack>
    </AppShell>
  );
}
