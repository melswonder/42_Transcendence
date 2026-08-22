import { redirect } from "next/navigation";
import { Group, Stack, Title } from "@mantine/core";

import { AppShell } from "@/components/app-shell";
import { ExportButtons } from "@/components/export-buttons";
import { MatchHistory } from "@/components/match-history";
import { PaginationControls } from "@/components/pagination-controls";
import { StatsFilters } from "@/components/stats-filters";
import { StatsStream } from "@/components/use-stats-stream";
import { getTranslations } from "next-intl/server";

import { getCurrentUser } from "@/lib/auth";
import { getMatches, type StatsFilters as Filters } from "@/lib/stats";

export default async function MatchesPage({
  searchParams,
}: {
  searchParams: Promise<Filters>;
}) {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  const filters = await searchParams;
  const matches = await getMatches(filters);
  const t = await getTranslations("matches");

  return (
    <AppShell user={user}>
      <StatsStream />
      <Stack gap="lg" maw={900} mx="auto">
        <Group justify="space-between" align="center">
          <Title order={2}>{t("title")}</Title>
          <ExportButtons />
        </Group>

        <div className="print:hidden">
          <StatsFilters />
        </div>

        <MatchHistory matches={matches.items} />

        <PaginationControls
          total={matches.total}
          limit={matches.limit}
          offset={matches.offset}
        />
      </Stack>
    </AppShell>
  );
}
