"use client";

import { DonutChart } from "@mantine/charts";
import { Group, Paper, Stack, Text, Title } from "@mantine/core";
import { useTranslations } from "next-intl";

import type { BreakdownSlice } from "@/lib/stats";

// 勝敗は意味の決まった色を当て、それ以外は順番に振る。
const OUTCOME_COLORS: Record<string, string> = {
  win: "emerald.5",
  loss: "red.6",
  draw: "gray.5",
};

const FALLBACK_COLORS = [
  "emerald.5",
  "cyan.5",
  "violet.5",
  "orange.5",
  "gray.5",
];

export function OutcomeChart({
  title,
  slices,
}: {
  title: string;
  slices: BreakdownSlice[];
}) {
  // API は機械可読な識別子を返し、画面の言葉はロケールごとの messages から引く。
  const tOutcomes = useTranslations("stats.outcomes");
  const tModes = useTranslations("stats.modes");
  const tResults = useTranslations("stats.results");
  const tStats = useTranslations("stats");
  const label = (key: string) => {
    if (tOutcomes.has(key)) return tOutcomes(key);
    if (tModes.has(key)) return tModes(key);
    if (tResults.has(key)) return tResults(key);
    return key;
  };
  const total = slices.reduce((sum, s) => sum + s.count, 0);

  if (total === 0) {
    return (
      <Paper p="lg">
        <Title order={3} size="h5" mb="md">
          {title}
        </Title>
        <Text c="dimmed" ta="center" py="xl">
          {tStats("emptyChart")}
        </Text>
      </Paper>
    );
  }

  const data = slices.map((slice, i) => ({
    name: label(slice.key),
    value: slice.count,
    color:
      OUTCOME_COLORS[slice.key] ?? FALLBACK_COLORS[i % FALLBACK_COLORS.length],
  }));

  return (
    <Paper p="lg">
      <Title order={3} size="h5" mb="md">
        {title}
      </Title>
      <Group justify="center" gap="xl" wrap="nowrap">
        <DonutChart
          data={data}
          size={170}
          thickness={26}
          tooltipDataSource="segment"
          chartLabel={tStats("totalGames", { count: total })}
        />
        <Stack gap={6}>
          {data.map((d) => (
            <Group key={d.name} gap="xs" wrap="nowrap">
              <span
                aria-hidden
                className="size-3 shrink-0 rounded-sm"
                style={{
                  background: `var(--mantine-color-${d.color.replace(".", "-")})`,
                }}
              />
              <Text size="sm">{d.name}</Text>
              <Text size="sm" c="dimmed" ff="monospace">
                {d.value}
              </Text>
            </Group>
          ))}
        </Stack>
      </Group>
    </Paper>
  );
}
