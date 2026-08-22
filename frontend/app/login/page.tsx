import Link from "next/link";
import { Alert, Anchor, Text } from "@mantine/core";
import { IconAlertCircle } from "@tabler/icons-react";
import { getTranslations } from "next-intl/server";

import { AuthCard } from "@/components/auth-card";
import { GoogleLoginButton } from "@/components/google-login-button";
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
    <AuthCard tagline={t("tagline")}>
      {message && (
        <Alert color="red" variant="light" icon={<IconAlertCircle />}>
          {message}
        </Alert>
      )}

      <GoogleLoginButton />

      <Text size="xs" c="dimmed" ta="center">
        {t("disclaimer")}
      </Text>

      <Text size="sm" ta="center">
        {t("noAccount")}{" "}
        <Anchor component={Link} href="/signup" size="sm">
          {t("signupLink")}
        </Anchor>
      </Text>
    </AuthCard>
  );
}
