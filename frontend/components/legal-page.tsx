"use client";

import Link from "next/link";
import { Anchor, Paper, Stack, Text, Title } from "@mantine/core";

export interface LegalSection {
  title: string;
  body: string;
}

/** Privacy Policy / Terms of Service 共通のドキュメントレイアウト。
 * 未ログインでも読めるよう、AppShell には載せない。
 */
export function LegalPage({
  title,
  updated,
  sections,
  backLabel,
}: {
  title: string;
  updated: string;
  sections: LegalSection[];
  backLabel: string;
}) {
  return (
    <main className="min-h-dvh bg-body p-4 md:p-10">
      <Paper p="xl" className="mx-auto w-full max-w-3xl">
        <Stack gap="lg">
          <Stack gap={4}>
            <Title order={1} size="h2">
              {title}
            </Title>
            <Text size="sm" c="dimmed">
              {updated}
            </Text>
          </Stack>

          {sections.map((section) => (
            <Stack key={section.title} gap={6}>
              <Title order={2} size="h4">
                {section.title}
              </Title>
              <Text size="sm" className="whitespace-pre-line">
                {section.body}
              </Text>
            </Stack>
          ))}

          <Anchor component={Link} href="/login" size="sm">
            {backLabel}
          </Anchor>
        </Stack>
      </Paper>
    </main>
  );
}
