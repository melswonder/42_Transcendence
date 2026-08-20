import { Alert, Paper, Stack, Text, ThemeIcon, Title } from "@mantine/core";
import { IconAlertCircle } from "@tabler/icons-react";

import { GoogleLoginButton } from "@/components/google-login-button";
import { QuoridorMark } from "@/components/quoridor-mark";
import { resolveLoginError } from "@/lib/login-error";

// searchParams は App Router では Promise で渡ってくるので await して取り出す。
export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;
  const message = resolveLoginError(error);

  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-4">
      <Paper p="xl" shadow="md" className="w-full max-w-md">
        <Stack gap="lg">
          <Stack align="center" gap="xs">
            <ThemeIcon size={64} radius="lg" variant="light">
              <QuoridorMark />
            </ThemeIcon>
            <Title order={1} className="text-3xl tracking-[0.15em]">
              TRANSCENDENCE
            </Title>
            <Text c="dimmed">Sign in to play</Text>
          </Stack>

          {message && (
            // Alert は既定で role="alert" を持つので、ここで付け直す必要はない。
            <Alert color="red" variant="light" icon={<IconAlertCircle />}>
              {message}
            </Alert>
          )}

          <GoogleLoginButton />

          <Text size="xs" c="dimmed" ta="center">
            続行すると、Google
            アカウントの表示名・メールアドレス・プロフィール画像を取得します。
          </Text>
        </Stack>
      </Paper>
    </main>
  );
}
