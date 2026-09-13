import { redirect } from "next/navigation";
import { Divider, Text } from "@mantine/core";
import { getTranslations } from "next-intl/server";

import { AuthCard } from "@/components/auth-card";
import { AuthLink } from "@/components/auth-link";
import { GoogleLoginButton } from "@/components/google-login-button";
import { PasswordAuthForm } from "@/components/password-auth-form";
import { getCurrentUser } from "@/lib/auth";

export default async function SignupPage() {
  const user = await getCurrentUser();
  if (user) redirect("/");

  const t = await getTranslations("signup");
  const tAuth = await getTranslations("auth");

  return (
    <AuthCard tagline={t("tagline")}>
      <PasswordAuthForm mode="signup" />

      <Divider label={tAuth("or")} labelPosition="center" />

      <GoogleLoginButton mode="signup" />

      <Text size="sm" ta="center">
        {t("haveAccount")} <AuthLink href="/login">{t("loginLink")}</AuthLink>
      </Text>
    </AuthCard>
  );
}
