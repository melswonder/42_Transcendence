"use client";

import { BarChart } from "@mantine/charts";
import { Paper, Title } from "@mantine/core";
import { useTranslations } from "next-intl";

import { EmptyChart } from "@/components/rating-chart";
import type { TimeseriesPoint } from "@/lib/stats";

/** 期間ごとの勝敗の内訳。積み上げにして「その日どれだけ遊んだか」も同時に見せる。 */
export function ActivityChart({ points }: { points: TimeseriesPoint[] }) {
  const t = useTranslations("stats");
  if (points.length === 0) {
    return <EmptyChart title={t("activityChart")} />;
  }

  return (
    <Paper p="lg">
      <Title order={3} size="h5" mb="md">
        {t("activityChart")}
      </Title>
      <BarChart
        h={260}
        data={points}
        dataKey="date"
        type="stacked"
        series={[
          { name: "wins", label: t("outcomes.win"), color: "emerald.5" },
          { name: "draws", label: t("outcomes.draw"), color: "gray.5" },
          { name: "losses", label: t("outcomes.loss"), color: "red.6" },
        ]}
        withLegend
        tooltipAnimationDuration={150}
      />
    </Paper>
  );
}
