import { Group, Paper, Stack, Text, ThemeIcon } from "@mantine/core";
import { useFormatter, useTranslations } from "next-intl";

import type { Match } from "@/lib/stats";

const OUTCOME_MARK: Record<string, { label: string; color: string }> = {
  win: { label: "W", color: "emerald" },
  loss: { label: "L", color: "red" },
  draw: { label: "D", color: "gray" },
};

/** 対戦履歴の一覧。1 件ずつ Paper を積むと境界線が二重になるので、
 * 外枠 1 枚の中に区切り線で並べる。
 */
export function MatchHistory({ matches }: { matches: Match[] }) {
  const t = useTranslations("matches");
  const tModes = useTranslations("stats.modes");
  const tResults = useTranslations("stats.results");
  // 日付・時刻はロケールの流儀で出す（2026/08/22 / Aug 22, 2026 / 22 août 2026）。
  const format = useFormatter();

  if (matches.length === 0) {
    return (
      <Paper p="xl">
        <Text c="dimmed" ta="center">
          {t("empty")}
        </Text>
      </Paper>
    );
  }

  return (
    <Paper className="overflow-hidden">
      {matches.map((match, i) => {
        const mark = OUTCOME_MARK[match.outcome];

        return (
          <Group
            key={match.id}
            justify="space-between"
            p="md"
            wrap="nowrap"
            className={i > 0 ? "border-t border-default-border" : undefined}
          >
            <Group gap="md" wrap="nowrap" className="min-w-0">
              <ThemeIcon
                size={40}
                radius="md"
                variant="light"
                color={mark.color}
              >
                <Text fw={700}>{mark.label}</Text>
              </ThemeIcon>
              <Stack gap={2} className="min-w-0">
                <Text fw={600} truncate>
                  {t("vs", { name: match.opponent.display_name })}
                </Text>
                <Text size="xs" c="dimmed">
                  {format.dateTime(new Date(match.finished_at), {
                    dateStyle: "medium",
                    timeStyle: "short",
                  })}{" "}
                  ・ {tModes.has(match.mode) ? tModes(match.mode) : match.mode}{" "}
                  ・{" "}
                  {tResults.has(match.result_type)
                    ? tResults(match.result_type)
                    : match.result_type}{" "}
                  ・ {t("moves", { count: match.total_moves })}
                </Text>
              </Stack>
            </Group>

            <Group gap="md" wrap="nowrap">
              <Text size="xs" c="dimmed" ff="monospace" visibleFrom="sm">
                {t("xp", { xp: match.xp_gained })}
              </Text>
              <Text
                ff="monospace"
                fw={500}
                c={ratingColor(match.rating_diff)}
                className="tabular-nums"
              >
                {formatDiff(match.rating_diff)}
              </Text>
            </Group>
          </Group>
        );
      })}
    </Paper>
  );
}

// レーティングが動かないモード（ランク戦以外）は 0 になるので、色を付けない。
function ratingColor(diff: number): string | undefined {
  if (diff > 0) return "emerald";
  if (diff < 0) return "red";

  return "dimmed";
}

function formatDiff(diff: number): string {
  if (diff === 0) return "±0";

  return diff > 0 ? `+${diff}` : `${diff}`;
}
