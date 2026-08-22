"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { Group, Paper, SegmentedControl, Select } from "@mantine/core";
import { useTranslations } from "next-intl";
import { DatePickerInput } from "@mantine/dates";
import { IconCalendar } from "@tabler/icons-react";

/** 絞り込みは URL のクエリに持つ。
 * Server Component が searchParams から読んでそのまま backend に渡せるので
 * 状態管理を増やさずに済み、共有やリロードにも耐える。
 */
export function StatsFilters({
  withInterval = false,
}: {
  withInterval?: boolean;
}) {
  const t = useTranslations("stats");
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const update = (patch: Record<string, string | null>) => {
    const params = new URLSearchParams(searchParams.toString());
    for (const [key, value] of Object.entries(patch)) {
      if (value) {
        params.set(key, value);
      } else {
        params.delete(key);
      }
    }
    // 条件を変えたら 1 ページ目に戻す。3 ページ目のまま絞ると空に見えるため。
    params.delete("offset");
    router.push(`${pathname}?${params.toString()}`);
  };

  const range: [string | null, string | null] = [
    searchParams.get("from"),
    searchParams.get("to"),
  ];

  return (
    <Paper p="md">
      <Group gap="md" align="flex-end" wrap="wrap">
        <DatePickerInput
          type="range"
          label={t("filters.period")}
          placeholder={t("filters.all")}
          leftSection={<IconCalendar size={16} />}
          value={range}
          onChange={([from, to]) => update({ from, to })}
          clearable
          miw={230}
        />

        <Select
          label={t("filters.mode")}
          placeholder={t("filters.all")}
          value={searchParams.get("mode")}
          onChange={(value) => update({ mode: value })}
          data={[
            { value: "ranked", label: t("modes.ranked") },
            { value: "casual", label: t("modes.casual") },
            { value: "ai", label: t("modes.ai") },
            { value: "friend", label: t("modes.friend") },
          ]}
          clearable
          w={150}
        />

        <Select
          label={t("filters.result")}
          placeholder={t("filters.all")}
          value={searchParams.get("outcome")}
          onChange={(value) => update({ outcome: value })}
          data={[
            { value: "win", label: t("outcomes.win") },
            { value: "loss", label: t("outcomes.loss") },
            { value: "draw", label: t("outcomes.draw") },
          ]}
          clearable
          w={130}
        />

        {withInterval && (
          <SegmentedControl
            value={searchParams.get("interval") ?? "day"}
            onChange={(value) => update({ interval: value })}
            data={[
              { value: "day", label: t("filters.daily") },
              { value: "week", label: t("filters.weekly") },
            ]}
          />
        )}
      </Group>
    </Paper>
  );
}
