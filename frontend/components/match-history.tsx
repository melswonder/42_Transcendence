import { Group, Paper, Stack, Text, ThemeIcon } from "@mantine/core";
import { IconChevronRight } from "@tabler/icons-react";

export interface Match {
  id: number;
  won: boolean;
  opponent: string;
  ratingDiff: number;
  moves: number;
  playedAt: string;
}

/** 対戦履歴の一覧。1 件ずつ Paper を積むと境界線が二重になるので、
 * 外枠 1 枚の中に区切り線で並べる。
 */
export function MatchHistory({ matches }: { matches: Match[] }) {
  if (matches.length === 0) {
    return (
      <Paper p="xl">
        <Text c="dimmed" ta="center">
          まだ対戦記録がありません。
        </Text>
      </Paper>
    );
  }

  return (
    <Paper className="overflow-hidden">
      {matches.map((match, i) => (
        <Group
          key={match.id}
          justify="space-between"
          p="md"
          className={i > 0 ? "border-t border-default-border" : undefined}
        >
          <Group gap="md">
            <ThemeIcon
              size={40}
              radius="md"
              variant="light"
              color={match.won ? "emerald" : "red"}
            >
              <Text fw={700}>{match.won ? "W" : "L"}</Text>
            </ThemeIcon>
            <Stack gap={2}>
              <Text fw={600}>vs {match.opponent}</Text>
              <Text size="xs" c="dimmed">
                {match.playedAt} ・ {match.moves} 手
              </Text>
            </Stack>
          </Group>

          <Group gap="md">
            <Text ff="monospace" fw={500} c={match.won ? "emerald" : "red"}>
              {match.ratingDiff > 0 ? `+${match.ratingDiff}` : match.ratingDiff}
            </Text>
            <IconChevronRight size={20} className="text-dimmed" />
          </Group>
        </Group>
      ))}
    </Paper>
  );
}
