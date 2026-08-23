import type { ReactNode } from "react";
import Image from "next/image";
import { Group, Stack, Title } from "@mantine/core";
import { useTranslations } from "next-intl";

import { LegalLinks } from "@/components/legal-links";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ShellFrame } from "@/components/shell-frame";
import { SideNav } from "@/components/side-nav";
import { UserCard } from "@/components/user-card";
import type { User } from "@/lib/auth";

/** ログイン後の共通の外枠。中身を組み立てるだけで、並べ方は ShellFrame が持つ。
 * Mantine の AppShell は header の高さ計算を持ち込む割に使わないので、素の flex で組む。
 */
export function AppShell({
  user,
  children,
}: {
  user: User;
  children: ReactNode;
}) {
  const t = useTranslations("common");

  const brand = (
    <Group gap="sm" wrap="nowrap">
      <Image
        src="/icon.png"
        alt=""
        width={28}
        height={28}
        className="rounded-lg"
      />
      <Title order={1} c="emerald" className="text-xl tracking-[0.2em]">
        {t("appName")}
      </Title>
    </Group>
  );

  // サイドバーにも Drawer にも同じものを出すので、要素として一度だけ組む。
  const nav = (
    <>
      <Stack gap={4} className="flex-1">
        <SideNav />
      </Stack>

      <LocaleSwitcher loggedIn />
      <UserCard user={user} />
      <LegalLinks />
    </>
  );

  return (
    <ShellFrame brand={brand} nav={nav}>
      {children}
    </ShellFrame>
  );
}
