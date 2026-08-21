import { Group, Paper, Stack, Text, ThemeIcon } from "@mantine/core";

import type { Match } from "@/lib/stats";

const MODE_LABELS: Record<string, string> = {
  ranked: "ランク戦",
  casual: "カジュアル",
  ai: "AI 対戦",
  friend: "フレンド戦",
};

const RESULT_LABELS: Record<string, string> = {
  goal: "ゴール到達",
  resign: "投了",
  timeout: "時間切れ",
  draw: "引き分け",
  abort: "中断",
};

const OUTCOME_MARK: Record<string, { label: string; color: string }> = {
  win: { label: "W", color: "emerald" },
  loss: { label: "L", color: "red" },
  draw: { label: "D", color: "gray" },
};

/** 対戦履歴の一覧。1 件ずつ Paper を積むと境界線が二重になるので、
 * 外枠 1 枚の中に区切り線で並べる。
 */
export function MatchHistory({ matches }: { matches: Match[] }) {
  if (matches.length === 0) {
    return (
      <Paper p="xl">
        <Text c="dimmed" ta="center">
          この条件に合う対戦がありません。
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
                  vs {match.opponent.display_name}
                </Text>
                <Text size="xs" c="dimmed">
                  {formatDateTime(match.finished_at)} ・{" "}
                  {MODE_LABELS[match.mode] ?? match.mode} ・{" "}
                  {RESULT_LABELS[match.result_type] ?? match.result_type} ・{" "}
                  {match.total_moves} 手
                </Text>
              </Stack>
            </Group>

            <Group gap="md" wrap="nowrap">
              <Text size="xs" c="dimmed" ff="monospace" visibleFrom="sm">
                +{match.xp_gained} XP
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

function formatDateTime(iso: string): string {
  // ロケール依存の揺れを避けるため、表示形式は固定する。
  const d = new Date(iso);
  const pad = (n: number) => `${n}`.padStart(2, "0");

  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
