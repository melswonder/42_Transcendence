"use client";

import { DonutChart } from "@mantine/charts";
import { Group, Paper, Stack, Text, Title } from "@mantine/core";

import type { BreakdownSlice } from "@/lib/stats";

// API が返す識別子を画面の言葉に直す。
// バックエンドが英語のまま返すのは、機械可読な値を API の正本にしておくため。
const LABELS: Record<string, string> = {
  win: "勝ち",
  loss: "負け",
  draw: "引き分け",
  goal: "ゴール到達",
  resign: "投了",
  timeout: "時間切れ",
  abort: "中断",
  ranked: "ランク戦",
  casual: "カジュアル",
  ai: "AI 対戦",
  friend: "フレンド戦",
};

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
  const total = slices.reduce((sum, s) => sum + s.count, 0);

  if (total === 0) {
    return (
      <Paper p="lg">
        <Title order={3} size="h5" mb="md">
          {title}
        </Title>
        <Text c="dimmed" ta="center" py="xl">
          この期間の対戦がありません。
        </Text>
      </Paper>
    );
  }

  const data = slices.map((slice, i) => ({
    name: LABELS[slice.key] ?? slice.key,
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
          chartLabel={`${total} 戦`}
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
