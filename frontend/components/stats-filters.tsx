"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { Group, Paper, SegmentedControl, Select } from "@mantine/core";
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
          label="期間"
          placeholder="すべて"
          leftSection={<IconCalendar size={16} />}
          value={range}
          onChange={([from, to]) => update({ from, to })}
          clearable
          miw={230}
        />

        <Select
          label="モード"
          placeholder="すべて"
          value={searchParams.get("mode")}
          onChange={(value) => update({ mode: value })}
          data={[
            { value: "ranked", label: "ランク戦" },
            { value: "casual", label: "カジュアル" },
            { value: "ai", label: "AI 対戦" },
            { value: "friend", label: "フレンド戦" },
          ]}
          clearable
          w={150}
        />

        <Select
          label="結果"
          placeholder="すべて"
          value={searchParams.get("outcome")}
          onChange={(value) => update({ outcome: value })}
          data={[
            { value: "win", label: "勝ち" },
            { value: "loss", label: "負け" },
            { value: "draw", label: "引き分け" },
          ]}
          clearable
          w={130}
        />

        {withInterval && (
          <SegmentedControl
            value={searchParams.get("interval") ?? "day"}
            onChange={(value) => update({ interval: value })}
            data={[
              { value: "day", label: "日別" },
              { value: "week", label: "週別" },
            ]}
          />
        )}
      </Group>
    </Paper>
  );
}
