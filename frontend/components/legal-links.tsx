"use client";

import Link from "next/link";
import { Anchor, Group } from "@mantine/core";
import { useTranslations } from "next-intl";

/** Privacy Policy / Terms へのフッターリンク。認証画面とサイドバーの両方に置く。 */
export function LegalLinks() {
  const t = useTranslations("legal");
  return (
    <Group justify="center" gap="md">
      <Anchor component={Link} href="/privacy" size="xs" c="dimmed">
        {t("footerPrivacy")}
      </Anchor>
      <Anchor component={Link} href="/terms" size="xs" c="dimmed">
        {t("footerTerms")}
      </Anchor>
    </Group>
  );
}
