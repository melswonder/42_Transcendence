import Image from "next/image";
import { Group, Stack, Text, Title } from "@mantine/core";
import {
  IconActivity,
  IconChevronRight,
  IconPlayerPlay,
  IconSearch,
} from "@tabler/icons-react";

import { AppShell } from "@/components/app-shell";
import { GameModeCard } from "@/components/game-mode-card";
import { LegalLinks } from "@/components/legal-links";
import { LinkButton } from "@/components/link-button";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { MatchHistory } from "@/components/match-history";
import { StatsStream } from "@/components/use-stats-stream";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";
import { getMatches } from "@/lib/stats";

// ホームに出す直近の対戦数。全部は /matches で見る。
const RECENT_LIMIT = 5;

export default async function HomePage() {
  const user = await getCurrentUser();
  if (!user) return <Landing />;
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

          <GameModeCard
            featured
            icon={<IconSearch size={24} />}
            title={t("quickMatch.title")}
            description={t("quickMatch.description")}
            href="/game"
          />
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

/** 未ログインで / に来た人向けの入口。ログインか新規登録かをここで選ぶ。 */
async function Landing() {
  const t = await getTranslations("landing");
  const tCommon = await getTranslations("common");

  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-4">
      <Stack align="center" gap="lg">
        <Image
          src="/icon.png"
          alt=""
          width={140}
          height={140}
          priority
          className="rounded-[2rem]"
        />
        <Stack align="center" gap={4}>
          <Title order={1} className="text-4xl tracking-[0.15em]">
            {tCommon("appName")}
          </Title>
          <Text c="dimmed">{t("tagline")}</Text>
        </Stack>
        <Group gap="sm">
          <LinkButton href="/login" size="md" variant="default" miw={140}>
            {t("login")}
          </LinkButton>
          <LinkButton href="/signup" size="md" miw={140}>
            {t("signup")}
          </LinkButton>
        </Group>
        <LocaleSwitcher />
        <LegalLinks />
      </Stack>
    </main>
  );
}
