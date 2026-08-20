import { Container, Stack, Text, ThemeIcon, Title } from "@mantine/core";
import { IconArrowRight } from "@tabler/icons-react";

import { LinkButton } from "@/components/link-button";
import { QuoridorMark } from "@/components/quoridor-mark";

export default function HomePage() {
  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-6">
      <Container size="sm">
        <Stack align="center" gap="xl">
          <ThemeIcon size={80} radius="lg" variant="light">
            <QuoridorMark size={44} />
          </ThemeIcon>

          <Stack align="center" gap="sm">
            <Title
              order={1}
              className="text-center text-4xl tracking-[0.2em] sm:text-5xl"
            >
              TRANSCENDENCE
            </Title>
            <Text size="lg" c="dimmed" ta="center" maw={420}>
              ブラウザで対戦できるコリドール。ログインしてゲームを始めましょう。
            </Text>
          </Stack>

          <LinkButton
            href="/login"
            size="lg"
            rightSection={<IconArrowRight size={18} />}
          >
            ログインして始める
          </LinkButton>
        </Stack>
      </Container>
    </main>
  );
}
