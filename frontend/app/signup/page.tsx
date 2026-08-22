import { redirect } from "next/navigation";
import { Divider, Text } from "@mantine/core";
import { getTranslations } from "next-intl/server";

import { AuthCard } from "@/components/auth-card";
import { AuthLink } from "@/components/auth-link";
import { GoogleLoginButton } from "@/components/google-login-button";
import { PasswordAuthForm } from "@/components/password-auth-form";
import { getCurrentUser } from "@/lib/auth";

/** サインアップ画面。認証は共通の Google OAuth で、初回ログイン時にアカウントが作られる。 */
export default async function SignupPage() {
  // ログイン済みならアカウントはもうあるので、ホームへ。
  const user = await getCurrentUser();
  if (user) redirect("/");

  const t = await getTranslations("signup");

  return (
    <AuthCard tagline={t("tagline")}>
      <PasswordAuthForm mode="signup" />

      <Divider label={t("or")} labelPosition="center" />

      <GoogleLoginButton mode="signup" />

      <Text size="xs" c="dimmed" ta="center">
        {t("disclaimer")}
      </Text>

      <Text size="sm" ta="center">
        {t("haveAccount")} <AuthLink href="/login">{t("loginLink")}</AuthLink>
      </Text>
    </AuthCard>
  );
}
