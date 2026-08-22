import { redirect } from "next/navigation";
import { Group, SimpleGrid, Stack, Title } from "@mantine/core";
import {
  IconActivity,
  IconChevronRight,
  IconPlayerPlay,
  IconRobot,
  IconSearch,
  IconUsers,
} from "@tabler/icons-react";

import { AppShell } from "@/components/app-shell";
import { GameModeCard } from "@/components/game-mode-card";
import { LinkButton } from "@/components/link-button";
import { MatchHistory } from "@/components/match-history";
import { StatsStream } from "@/components/use-stats-stream";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";
import { getMatches } from "@/lib/stats";

// ホームに出す直近の対戦数。全部は /matches で見る。
const RECENT_LIMIT = 5;

export default async function HomePage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");
  const t = await getTranslations("home");

  const recent = await getMatches({ limit: `${RECENT_LIMIT}` });

  return (
    <AppShell user={user}>
      <StatsStream />
      <Stack gap="xl" maw={900} mx="auto">
        <section>
          <Group gap="xs" mb="lg">
            <IconPlayerPlay size={24} className="text-emerald-500" />
            <Title order={2} size="h3">
              {t("newMatch")}
            </Title>
          </Group>

          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
            <GameModeCard
              featured
              icon={<IconSearch size={24} />}
              title={t("quickMatch.title")}
              description={t("quickMatch.description")}
              href="/game"
            />
            <GameModeCard
              icon={<IconRobot size={24} />}
              title={t("vsAi.title")}
              description={t("vsAi.description")}
              comingSoon
            />
            <GameModeCard
              icon={<IconUsers size={24} />}
              title={t("vsFriend.title")}
              description={t("vsFriend.description")}
              comingSoon
            />
          </SimpleGrid>
        </section>

        <section>
          <Group justify="space-between" mb="lg">
            <Group gap="xs">
              <IconActivity size={24} className="text-dimmed" />
              <Title order={2} size="h3">
                最近の対戦
              </Title>
            </Group>
            <LinkButton
              href="/matches"
              variant="subtle"
              size="xs"
              rightSection={<IconChevronRight size={16} />}
            >
              すべて見る
            </LinkButton>
          </Group>

          <MatchHistory matches={recent.items} />
        </section>
      </Stack>
    </AppShell>
  );
}
