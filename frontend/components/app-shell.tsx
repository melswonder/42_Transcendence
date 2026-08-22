import type { ReactNode } from "react";
import { ScrollArea, Stack, Title } from "@mantine/core";

import { LegalLinks } from "@/components/legal-links";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { SideNav } from "@/components/side-nav";
import { UserCard } from "@/components/user-card";
import type { User } from "@/lib/auth";

/** ログイン後の共通の外枠。左にサイドバー、右に各画面の中身。
 * Mantine の AppShell は header の高さ計算を持ち込む割に使わないので、素の flex で組む。
 */
export function AppShell({
  user,
  children,
}: {
  user: User;
  children: ReactNode;
}) {
  return (
    <div className="flex min-h-dvh flex-col bg-body md:flex-row">
      <aside className="flex shrink-0 flex-col gap-4 border-default-border bg-dark-600 p-4 max-md:border-b md:w-64 md:border-r">
        <Title
          order={1}
          c="emerald"
          className="px-2 py-3 text-xl tracking-[0.2em]"
        >
          QUORIDOR
        </Title>

        <Stack gap={4} className="flex-1">
          <SideNav />
        </Stack>

        <LocaleSwitcher loggedIn />
        <UserCard user={user} />
        <LegalLinks />
      </aside>

      <ScrollArea className="flex-1" scrollbars="y">
        <div className="p-6 md:p-10">{children}</div>
      </ScrollArea>
    </div>
  );
}
