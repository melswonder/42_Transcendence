import { Alert, Paper, Stack, Text, ThemeIcon, Title } from "@mantine/core";
import { IconAlertCircle } from "@tabler/icons-react";

import { GoogleLoginButton } from "@/components/google-login-button";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { QuoridorMark } from "@/components/quoridor-mark";
import { getTranslations } from "next-intl/server";

import { resolveLoginErrorKey } from "@/lib/login-error";

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;
  const t = await getTranslations("login");
  const errorKey = resolveLoginErrorKey(error);
  const message = errorKey ? t(`errors.${errorKey}`) : null;

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
            <Text c="dimmed">{t("tagline")}</Text>
          </Stack>

          {message && (
            <Alert color="red" variant="light" icon={<IconAlertCircle />}>
              {message}
            </Alert>
          )}

          <GoogleLoginButton />

          <Text size="xs" c="dimmed" ta="center">
            {t("disclaimer")}
          </Text>

          <LocaleSwitcher />
        </Stack>
      </Paper>
    </main>
  );
}
