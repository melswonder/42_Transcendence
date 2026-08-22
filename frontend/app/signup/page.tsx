import { redirect } from "next/navigation";
import { Text } from "@mantine/core";
import { getTranslations } from "next-intl/server";

import { AuthCard } from "@/components/auth-card";
import { AuthLink } from "@/components/auth-link";
import { GoogleLoginButton } from "@/components/google-login-button";
import { getCurrentUser } from "@/lib/auth";

/** サインアップ画面。認証は Google OAuth で、初回ログイン時に
 * アカウントが自動で作られる。ログイン画面とは文言と導線だけが違う。
 */
export default async function SignupPage() {
  // ログイン済みならアカウントはもうあるので、ホームへ。
  const user = await getCurrentUser();
  if (user) redirect("/");

  const t = await getTranslations("signup");

  return (
    <AuthCard tagline={t("tagline")}>
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
