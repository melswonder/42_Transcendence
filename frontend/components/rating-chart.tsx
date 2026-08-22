"use client";

import { LineChart } from "@mantine/charts";
import { Paper, Text, Title } from "@mantine/core";
import { useTranslations } from "next-intl";

import type { TimeseriesPoint } from "@/lib/stats";

/** レーティングの推移。値は各コマの最後の対戦を終えた時点。 */
export function RatingChart({ points }: { points: TimeseriesPoint[] }) {
  const t = useTranslations("stats");
  if (points.length === 0) {
    return <EmptyChart title={t("ratingChart")} />;
  }

  return (
    <Paper p="lg">
      <Title order={3} size="h5" mb="md">
        {t("ratingChart")}
      </Title>
      <LineChart
        h={260}
        data={points}
        dataKey="date"
        series={[
          { name: "rating", label: t("tiles.rating"), color: "emerald.5" },
        ]}
        curveType="monotone"
        withDots={points.length <= 30}
        // 1200 前後の変動を見たいので、0 起点にせず値域に合わせる。
        yAxisProps={{ domain: ["dataMin - 30", "dataMax + 30"] }}
        tooltipAnimationDuration={150}
      />
    </Paper>
  );
}

export function EmptyChart({ title }: { title: string }) {
  const t = useTranslations("stats");
  return (
    <Paper p="lg">
      <Title order={3} size="h5" mb="md">
        {title}
      </Title>
      <Text c="dimmed" ta="center" py="xl">
        {t("emptyChart")}
      </Text>
    </Paper>
  );
}
