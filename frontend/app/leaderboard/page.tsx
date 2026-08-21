import { redirect } from "next/navigation";
import { Avatar, Group, Paper, Stack, Text, Title } from "@mantine/core";

import { AppShell } from "@/components/app-shell";
import { PaginationControls } from "@/components/pagination-controls";
import { StatsStream } from "@/components/use-stats-stream";
import { getCurrentUser } from "@/lib/auth";
import {
  getLeaderboard,
  type LeaderboardEntry,
  type StatsFilters,
} from "@/lib/stats";

export default async function LeaderboardPage({
  searchParams,
}: {
  searchParams: Promise<StatsFilters>;
}) {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  const filters = await searchParams;
  const board = await getLeaderboard(filters);

  // 自分が表示範囲に居るかどうかで、下の「自分の順位」を出し分ける。
  const meInPage = board.items.some((entry) => entry.user.id === user.id);

  return (
    <AppShell user={user}>
      <StatsStream />
      <Stack gap="lg" maw={800} mx="auto">
        <Title order={2}>ランキング</Title>

        <Paper className="overflow-hidden">
          {board.items.map((entry, i) => (
            <LeaderboardRow
              key={entry.user.id}
              entry={entry}
              isMe={entry.user.id === user.id}
              divided={i > 0}
            />
          ))}
        </Paper>

        {board.me && !meInPage && (
          <Stack gap={4}>
            <Text size="sm" c="dimmed">
              自分の順位
            </Text>
            <Paper className="overflow-hidden">
              <LeaderboardRow entry={board.me} isMe divided={false} />
            </Paper>
          </Stack>
        )}

        <PaginationControls
          total={board.total}
          limit={board.limit}
          offset={board.offset}
        />
      </Stack>
    </AppShell>
  );
}

function LeaderboardRow({
  entry,
  isMe,
  divided,
}: {
  entry: LeaderboardEntry;
  isMe: boolean;
  divided: boolean;
}) {
  return (
    <Group
      justify="space-between"
      p="md"
      wrap="nowrap"
      className={[
        divided ? "border-t border-default-border" : "",
        isMe ? "bg-emerald-900/40" : "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <Group gap="md" wrap="nowrap" className="min-w-0">
        <Text
          ff="monospace"
          fw={700}
          size="lg"
          w={44}
          ta="right"
          c={entry.rank <= 3 ? "emerald" : "dimmed"}
        >
          {entry.rank}
        </Text>
        <Avatar radius="xl" color="emerald" variant="filled" size="md">
          {entry.user.display_name.charAt(0).toUpperCase()}
        </Avatar>
        <Stack gap={0} className="min-w-0">
          <Text fw={600} truncate>
            {entry.user.display_name}
          </Text>
          <Text size="xs" c="dimmed" truncate>
            @{entry.user.handle} ・ Lv.{entry.user.level}
          </Text>
        </Stack>
      </Group>

      <Group gap="lg" wrap="nowrap">
        <Stack gap={0} align="flex-end" visibleFrom="xs">
          <Text size="sm">
            {entry.wins}勝 {entry.losses}敗
          </Text>
          <Text size="xs" c="dimmed">
            勝率 {(entry.win_rate * 100).toFixed(0)}%
          </Text>
        </Stack>
        <Text ff="monospace" fw={700} size="lg" className="tabular-nums">
          {entry.rating}
        </Text>
      </Group>
    </Group>
  );
}
