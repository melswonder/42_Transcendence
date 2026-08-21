import { redirect } from "next/navigation";
import { Group, SimpleGrid, Stack, Title } from "@mantine/core";
import {
  IconActivity,
  IconPlayerPlay,
  IconRobot,
  IconSearch,
  IconUsers,
} from "@tabler/icons-react";

import { AppShell } from "@/components/app-shell";
import { GameModeCard } from "@/components/game-mode-card";
import { MatchHistory, type Match } from "@/components/match-history";
import { getCurrentUser } from "@/lib/auth";

// 対戦機能がまだ無いので仮の値。API ができたら差し替える。
const matches: Match[] = [
  {
    id: 1,
    won: true,
    opponent: "Alex99",
    ratingDiff: 12,
    moves: 42,
    playedAt: "2 時間前",
  },
  {
    id: 2,
    won: false,
    opponent: "GrandMasterQ",
    ratingDiff: -8,
    moves: 65,
    playedAt: "5 時間前",
  },
  {
    id: 3,
    won: true,
    opponent: "Rookie_1",
    ratingDiff: 10,
    moves: 28,
    playedAt: "昨日",
  },
];

export default async function HomePage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return (
    <AppShell user={user}>
      <Stack gap="xl" maw={900} mx="auto">
        <section>
          <Group gap="xs" mb="lg">
            <IconPlayerPlay size={24} className="text-emerald-500" />
            <Title order={2} size="h3">
              新しい対戦
            </Title>
          </Group>

          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
            <GameModeCard
              featured
              icon={<IconSearch size={24} />}
              title="クイックマッチ"
              description="同じくらいの実力の相手とランク戦を始めます。"
              href="/game"
            />
            <GameModeCard
              icon={<IconRobot size={24} />}
              title="コンピュータ対戦"
              description="AI を相手に練習します。"
              comingSoon
            />
            <GameModeCard
              icon={<IconUsers size={24} />}
              title="フレンド対戦"
              description="合言葉を使って部屋を作ります。"
              comingSoon
            />
          </SimpleGrid>
        </section>

        <section>
          <Group gap="xs" mb="lg">
            <IconActivity size={24} className="text-dimmed" />
            <Title order={2} size="h3">
              最近の対戦
            </Title>
          </Group>

          <MatchHistory matches={matches} />
        </section>
      </Stack>
    </AppShell>
  );
}
