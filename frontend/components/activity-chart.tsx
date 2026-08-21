"use client";

import { BarChart } from "@mantine/charts";
import { Paper, Title } from "@mantine/core";

import { EmptyChart } from "@/components/rating-chart";
import type { TimeseriesPoint } from "@/lib/stats";

/** 期間ごとの勝敗の内訳。積み上げにして「その日どれだけ遊んだか」も同時に見せる。 */
export function ActivityChart({ points }: { points: TimeseriesPoint[] }) {
  if (points.length === 0) {
    return <EmptyChart title="対戦数の推移" />;
  }

  return (
    <Paper p="lg">
      <Title order={3} size="h5" mb="md">
        対戦数の推移
      </Title>
      <BarChart
        h={260}
        data={points}
        dataKey="date"
        type="stacked"
        series={[
          { name: "wins", label: "勝ち", color: "emerald.5" },
          { name: "draws", label: "引き分け", color: "gray.5" },
          { name: "losses", label: "負け", color: "red.6" },
        ]}
        withLegend
        tooltipAnimationDuration={150}
      />
    </Paper>
  );
}
